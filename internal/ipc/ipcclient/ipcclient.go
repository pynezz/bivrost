package ipcclient

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/ipc"
	"github.com/pynezz/bivrost/internal/util"
)

func countDown(secLeft int) {
	for i := secLeft; i > 0; i-- {
		i--
		time.Sleep(time.Second)
		util.PrintInfo(util.Overwrite + strconv.Itoa(i) + " seconds left" + util.Backspace)
		countDown(i)
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
}

// Set description with format string for easier type conversion
func (c *IPCClient) SetDescf(desc string, args ...interface{}) {
	c.Desc = fmt.Sprintf(desc, args...)
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

// We'll get the path from the config, but for now let's just hard code it
func (c *IPCClient) Connect(module string) error {
	tmpDir := os.TempDir()
	bivrostTmpDir := path.Join(tmpDir, "bivrost")          // temp directory
	socketPath := path.Join(bivrostTmpDir, module+".sock") // full path + filename
	c.Name = module
	c.SetDescf("IPC testing client for %s", module)

	// Check if the socket exists
	exist := socketExists(socketPath)
	if !exist {
		fmt.Println("Socket not found.")
		util.PrintColorUnderline(util.DarkYellow, "Retry? [Y/n]")
		var response string
		fmt.Scanln(&response)
		if response[0] == 'n' {
			return fmt.Errorf("socket not found")
		}
		c.Connect(module)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("Dial error:", err)
		return err
	}
	c.conn = conn
	c.Identifier = ipc.IDENTIFIERS["threat_intel"] // Should equal to  0x54, 0x48, 0x52, 0x49

	util.PrintColorAndBg(util.BgGray, util.BgCyan, "Connected to "+socketPath)

	// Print box with client info
	util.PrintColor(util.Cyan, c.Stringify())

	// Add the connection to the client buffer // NB! We may not need this
	// clientsBuffer[module] = c
	return nil
}

func (c *IPCClient) NewIPCMessage(message string, t ipc.MsgType) (*ipc.IPCRequest, error) {
	checksum := crc32.ChecksumIEEE([]byte(message))
	return &ipc.IPCRequest{
		Header: ipc.IPCHeader{
			Identifier:  c.Identifier,
			MessageType: byte(t),
		},
		Message: ipc.IPCMessage{
			Data:       []byte(message),
			StringData: message,
		},
		Checksum32: int(checksum),
	}, nil
}

func userRetry() bool {
	fmt.Println("IPCClient not connected\nDid you forget to call Connect()?")
	util.PrintWarning("Do you want to retry? [Y/n]")

	var retry string
	fmt.Scanln(&retry)
	return retry[0] != 'n' // If the user doesn't want to retry, return false
}

// sendMessage sends a string message to the connection
func (c *IPCClient) SendMessage(message string) error {
	if c.conn == nil {
		if !userRetry() {
			return fmt.Errorf("connection not established")
		} else {
			c.Connect(message)
		}
	}
	_, err := c.conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Write error:", err)
		return err
	}
	fmt.Println("Message sent:", message)

	return nil
}

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

	return nil
}

/*
NewMessage creates a new IPC message

message string: The message to send

t: The message type
*/
func (c *IPCClient) CreateReq(message string, t ipc.MsgType) *ipc.IPCRequest {
	checksum := crc32.ChecksumIEEE([]byte(message))
	util.PrintDebug("Created IPC checksum: " + strconv.Itoa(int(checksum)))

	return &ipc.IPCRequest{
		MessageSignature: ipc.IPCID,
		Header: ipc.IPCHeader{
			Identifier:  c.Identifier,
			MessageType: byte(t),
		},
		Message: ipc.IPCMessage{
			Data:       []byte(message),
			StringData: message,
		},
		Checksum32: int(checksum),
	}
}

// Close the connection
func (c *IPCClient) Close() {
	c.conn.Close()
}
