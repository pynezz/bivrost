package ipcclient

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/ipc"
	"github.com/pynezz/bivrost/internal/util"
)

type IPCClient struct {
	Name string // Name of the module
	Desc string // Description of the module

	Identifier [4]byte // Identifier of the module

	Sock string   // Path to the UNIX domain socket
	conn net.Conn // Connection to the IPC server (UNIX domain socket)
}

func countDown(secLeft int) { // i--
	util.PrintInfo(util.Overwrite + strconv.Itoa(secLeft) + " seconds left" + util.Backspace)
	time.Sleep(time.Second)
	if secLeft > 0 {
		countDown(secLeft - 1)
	}
}

func socketExists(socketPath string) bool {
	if !fsutil.FileExists(socketPath) {
		util.PrintError("The UNIX domain socket does not exist")
		util.PrintInfo("Retrying in 5 seconds...")
		countDown(5)
		return false
	}
	return true
}

func NewIPCClient() *IPCClient {
	return &IPCClient{}
}

func init() {
	gob.Register(&ipc.IPCRequest{})
	gob.Register(&ipc.IPCMessage{})
	gob.Register(&ipc.IPCHeader{})
	gob.Register(&ipc.IPCMessageId{})
	gob.Register(&ipc.IPCResponse{})
}

// Set description with format string for easier type conversion
func (c *IPCClient) SetDescf(desc string, args ...interface{}) {
	c.Desc = fmt.Sprintf(desc, args...)
}

func (c *IPCClient) Stringify() string {
	if c.Name == "" {
		util.PrintWarning("No name set for IPCClient")
		c.Name = "IPCClient"
	}
	if c.Desc == "" {
		util.PrintWarning("No description set for IPCClient")
		c.Desc = "IPC testing client"
	}
	if c.Identifier == [4]byte{} {
		util.PrintWarning("No identifier set for IPCClient")
		c.Identifier = ipc.IDENTIFIERS["threat_intel"]
	}

	stringified := fmt.Sprintln("IPCCLIENT")
	stringified += fmt.Sprintln("-----------")
	stringified += fmt.Sprintf("Name:        %s\n", c.Name)
	stringified += fmt.Sprintf("Description: %s\n", c.Desc)
	stringified += fmt.Sprintf("Identifier:  %s\n", c.Identifier)

	return util.FormatRoundedBox(stringified)
}

// returns a bool (retry) and an error
func existHandler(exist bool) (bool, error) {
	if !exist {
		util.PrintError("socket (" + DefaultSocketPath() + ") not found")
		util.PrintColorUnderline(util.DarkYellow, "Retry? [Y/n]")
		var response string
		fmt.Scanln(&response)
		if response[0] == 'n' {
			return false, fmt.Errorf("socket not found")
		}
		return true, nil
	}
	return false, nil
}

func (c *IPCClient) SetSocket(socketPath string) error {
	if socketPath == "" {
		socketPath = DefaultSocketPath()
	}
	c.Sock = socketPath

	retry, err := existHandler(socketExists(socketPath))
	if err != nil {
		return err
	}
	if retry {
		c.SetSocket(socketPath)
	}
	return err
}

// Get the default socket path (UNIX domain socket path, /tmp/bivrost/bivrost.sock)
func DefaultSocketPath() string {
	tmpDir := os.TempDir() // Get the temporary directory, OS agnostic
	bivrostTmpDir := path.Join(tmpDir, "bivrost")
	return path.Join(bivrostTmpDir, "bivrost.sock")
}

// We'll get the path from the config, but for now let's just hard code it
func (c *IPCClient) Connect(module string) error {
	c.SetDescf("IPC testing client for %s", module)
	c.Name = module

	// Check if the socket exists
	err := c.SetSocket(DefaultSocketPath())
	if err != nil { // The socket did not exist and the user did not want to retry
		return err // Return the error
	}

	conn, err := net.Dial("unix", c.Sock)
	if err != nil {
		fmt.Println("Dial error:", err)
		return err
	}
	c.conn = conn
	c.Identifier = ipc.IDENTIFIERS["threat_intel"] // Should equal to 0x54, 0x48, 0x52, 0x49,

	util.PrintColorAndBg(util.BgGray, util.BgCyan, "Connected to "+c.Sock)

	// Print box with client info
	util.PrintColor(util.Cyan, c.Stringify())

	return nil
}

// userRetry asks the user if they want to retry connecting to the IPC server.
func userRetry() bool {
	fmt.Println("IPCClient not connected\nDid you forget to call Connect()?")
	util.PrintWarning("Do you want to retry? [Y/n]")

	var retry string
	fmt.Scanln(&retry)
	return retry[0] != 'n' // If the user doesn't want to retry, return false
}

func (c *IPCClient) AwaitResponse() error {
	var err error

	if c.conn == nil {
		util.PrintError("Connection not established")
	}

	req, err := parseConnection(c.conn)
	if err != nil {
		if err.Error() == "EOF" {
			util.PrintWarning("Client disconnected")
			return err
		}
		util.PrintError("Error parsing the connection")
		return err
	}
	util.PrintSuccess("Received response from server: " + req.Message.StringData)

	if string(req.Message.Data) == "OK" {
		util.PrintColorf(util.LightCyan, "Message type: %v\n", req.Header.MessageType)
		util.PrintSuccess("Checksums match")
	} else {
		util.PrintError("Checksums do not match")
		return fmt.Errorf("checksums do not match")
	}

	return nil
}

// SendIPCMessage sends an IPC message to the server.
func (c *IPCClient) SendIPCMessage(msg *ipc.IPCRequest) error {
	var bBuffer bytes.Buffer
	encoder := gob.NewEncoder(&bBuffer)
	err := encoder.Encode(msg)
	if err != nil {
		return err
	}

	if c.conn == nil {
		if !userRetry() {
			return fmt.Errorf("connection not established")
		} else {
			c.Connect("bivrost")
		}
	}

	util.PrintItalic("Sending encoded message to server...")
	_, err = c.conn.Write(bBuffer.Bytes())
	if err != nil {
		fmt.Println("Write error:", err)
		return err
	}
	util.PrintSuccess("Message sent: " + msg.Message.StringData)

	util.PrintDebug("Awaiting response...")
	err = c.AwaitResponse()
	if err != nil {
		util.PrintError("Error receiving response from server")
		fmt.Println(err)
		return err
	}

	return nil
}

// NewMessage creates a new IPC message.
func (c *IPCClient) CreateReq(message string, t ipc.MsgType, dataType ipc.DataType) *ipc.IPCRequest {
	checksum := crc32.ChecksumIEEE([]byte(message))
	util.PrintDebug("Created IPC checksum: " + strconv.Itoa(int(checksum)))

	return &ipc.IPCRequest{
		MessageSignature: ipc.IPCID,
		Header: ipc.IPCHeader{
			Identifier:  c.Identifier,
			MessageType: byte(t),
		},
		Message: ipc.IPCMessage{
			Datatype:   dataType,
			Data:       []byte(message),
			StringData: message,
		},
		Timestamp:  util.UnixNanoTimestamp(),
		Checksum32: int(checksum),
	}
}

func (c *IPCClient) CreateGenericReq(message interface{}, t ipc.MsgType, dataType ipc.DataType) *ipc.IPCRequest {
	util.PrintDebug("[CLIENT] Creating a generic IPC request...")
	var data []byte
	var err error

	switch dataType {
	case ipc.DATA_TEXT:
		data = []byte(message.(string))
	case ipc.DATA_INT:
		data = []byte(strconv.Itoa(message.(int)))
	case ipc.DATA_JSON:
		data, err = json.Marshal(message)
		if err != nil {
			// Handle the error
			util.PrintError("[CLIENT] Error marshaling JSON data:" + err.Error())
			return nil
		}
		util.PrintDebug("[CLIENT] Marshaling JSON data...")

	case ipc.DATA_YAML:
		fmt.Println("[CLIENT] Marshaling YAML data...")
		data, err = yaml.Marshal(message)
		if err != nil {
			util.PrintError("[CLIENT] Error marshaling YAML data:" + err.Error())
			return nil
		}
	case ipc.DATA_BIN:
		data = message.([]byte)
	}

	checksum := crc32.ChecksumIEEE(data)
	util.PrintDebug("[CLIENT] Created IPC checksum: " + strconv.Itoa(int(checksum)))

	return &ipc.IPCRequest{
		MessageSignature: ipc.IPCID,
		Header: ipc.IPCHeader{
			Identifier:  c.Identifier,
			MessageType: byte(t),
		},
		Message: ipc.IPCMessage{
			Datatype:   dataType,
			Data:       data,
			StringData: fmt.Sprintf("%v", message),
		},
		Timestamp:  util.UnixNanoTimestamp(),
		Checksum32: int(checksum),
	}
}

// Return the parsed IPCRequest object
func parseConnection(c net.Conn) (ipc.IPCRequest, error) {
	var request ipc.IPCRequest
	// var reqBuffer bytes.Buffer

	util.PrintDebug("[CLIENT] Trying to decode the bytes to a request struct...")
	util.PrintColorf(util.LightCyan, "[CLIENT] Decoding the bytes to a request struct... %v", c)

	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&request)
	if err != nil {
		if err.Error() == "EOF" {
			util.PrintWarning("parseConnection: EOF error, connection closed")
			return request, err
		}
		util.PrintWarning("parseConnection: Error decoding the request \n > " + err.Error())
		return request, err
	}

	util.PrintDebug("Trying to encode the bytes to a request struct...")
	fmt.Println(request.Stringify())
	util.PrintDebug("--------------------")

	util.PrintSuccess("[ipcclient.go] Parsed the message signature!")
	fmt.Printf("Message ID: %v\n", request.MessageSignature)

	return request, nil
}

// Close the connection
func (c *IPCClient) Close() {
	c.conn.Close()
}

// Get a key from a value in a map such as IDENTIFIERS
func getKeyFromValue(value [4]byte) (string, bool) {
	for key, val := range ipc.IDENTIFIERS {
		if val == value {
			return key, true
		}
	}
	return "", false
}
