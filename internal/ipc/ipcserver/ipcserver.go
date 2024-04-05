package ipcserver

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/pynezz/bivrost/internal/ipc"
	"github.com/pynezz/bivrost/internal/util"
	"gopkg.in/yaml.v3"
)

/* CONSTANTS
 * The following constants are used for the IPC communication between the connector and the other modules.
 */
const (
	AF_UNIX  = "unix"     // UNIX domain sockets
	AF_DGRAM = "unixgram" // UNIX domain datagram sockets as specified in net package
)

/* IDENTIFIERS
 * To identify the module, the client will send a 4 byte identifier as part of the header.
 */
var IDENTIFIERS map[string][]byte

var (
	THREAT_INTEL = []byte{0x54, 0x48, 0x52, 0x49} // "THRI"
	BIVROST      = []byte{0x42, 0x49, 0x56, 0x52} // "BIVR"
)

var MSGTYPE map[string]byte

/* TYPES
 * Types for the IPC communication between the connector and the other modules.
 */
type IPCServer struct {
	path string

	conn net.Listener
}

// (i) vv Highly likely not needed!->
type Connections struct {
	sockets map[string]*IPCServer
}

var connections Connections

// (i) ^^ likely not needed ^^ -- <-

func init() {
	IDENTIFIERS = map[string][]byte{
		"threat_intel": THREAT_INTEL,
		"bivrost":      BIVROST,
	}

	gob.Register(ipc.IPCRequest{})
	gob.Register(ipc.IPCMessage{})
	gob.Register(ipc.IPCHeader{})
	gob.Register(ipc.IPCMessageId{})
	gob.Register(&ipc.IPCResponse{})
}

func NewIPCServer(path string) *IPCServer {
	return &IPCServer{
		path: path,
	}
}

// Write a socket file and add it to the map
func (s *IPCServer) InitServerSocket() bool {
	// Making sure the socket is clean before starting
	if err := os.RemoveAll(s.path); err != nil {
		util.PrintError("InitServerSocket(): Failed to remove old socket: " + err.Error())
		return false
	}

	return true
}

// Creates a new listener on the socket path (which should be set in the config in the future)
func (s *IPCServer) Listen() {
	util.PrintColorBold(util.DarkGreen, "🎉 IPC server running!")
	s.conn, _ = net.Listen(AF_UNIX, s.path)
	util.PrintColorf(util.LightCyan, "[🔌SOCKETS] Starting listener on %s", s.path)

	for {
		util.PrintDebug("Waiting for connection...")
		conn, err := s.conn.Accept()
		util.PrintColorf(util.LightCyan, "[🔌SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			util.PrintError("Listen(): " + err.Error())
			continue
		}

		handleConnection(conn)
	}
}

func Cleanup() {
	// for _, server := range connections.sockets {
	// 	util.PrintItalic("Cleaning up IPC server: " + server.path)
	// 	err := os.Remove(server.path)
	// 	if err != nil {
	// 		util.PrintError("Cleanup(): " + err.Error())
	// 	}

	// 	util.PrintItalic("Closing connection: " + server.conn.Addr().String())
	// 	server.conn.Close()
	// }
	util.PrintItalic("\t... IPC server cleanup complete.")
}

func crc(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
	// s, err := c.Write(b)
	// if err != nil {
	// 	util.PrintError("crc(): " + err.Error())
	// 	return 0xFF // Error
	// }
	// return s
}

// Function to create a new IPCMessage based on the identifier key
func NewIPCMessage(identifierKey string, messageType byte, data []byte) (*ipc.IPCRequest, error) {
	identifier, ok := IDENTIFIERS[identifierKey]
	if !ok {
		return nil, fmt.Errorf("invalid identifier key: %s", identifierKey)
	}

	var id [4]byte
	copy(id[:], identifier[:4]) // Ensure no out of bounds panic

	crcsum32 := crc(data)

	message := ipc.IPCMessage{
		Data:       data,
		StringData: string(data),
	}

	return &ipc.IPCRequest{
		Header: ipc.IPCHeader{
			Identifier:  id,
			MessageType: messageType,
		},
		Message:    message,
		Timestamp:  util.UnixNanoTimestamp(),
		Checksum32: int(crcsum32),
	}, nil
}

func newIPCResponse(req ipc.IPCRequest, success bool, message string) *ipc.IPCResponse {
	return &ipc.IPCResponse{
		Request:    req,
		Success:    success,
		Message:    message,
		Checksum32: int(req.Checksum32),
	}
}

// Check the message signature (the first 6 bytes) to see if it's a valid IPC message
func parseMessageSignature(message []byte) (ipc.IPCMessageId, error) {
	util.PrintWarning("Possibly error and bug prone section ahead! 🐞")
	util.PrintWarning("Will try to parse the bytes of the message to see if it's a valid IPC message... 🧐")

	if len(message) < 6 {
		return nil, fmt.Errorf("parseMessageSignature: Message too short")
	}

	var bBuffer bytes.Buffer
	decoder := gob.NewDecoder(&bBuffer)
	err := decoder.Decode(&message)
	if err != nil {
		return nil, err
	}

	util.PrintDebug("Decoded message: \n " + bBuffer.String())

	var id ipc.IPCMessageId
	// Convert bBuffer to a byte slice before slicing it
	copy(id[:], bBuffer.Bytes()[:len(ipc.IPCID)])

	util.PrintWarning("Attempting to compare the message signature...")

	if bytes.Equal(id, ipc.IPCID) {
		util.PrintSuccess("Message signature is valid!")
		util.PrintDebug("Message signature: " + string(id))
		return id, nil
	}

	util.PrintError("Message signature is invalid!")
	util.PrintError("Message signature: " + string(id))
	return nil, fmt.Errorf("parseMessageSignature: Invalid message signature")
}

// Return the parsed IPCRequest object
func parseConnection(c net.Conn) (ipc.IPCRequest, error) {
	var request ipc.IPCRequest
	// var reqBuffer bytes.Buffer

	util.PrintDebug("Trying to decode the bytes to a request struct...")
	util.PrintColorf(util.LightCyan, "Decoding the bytes to a request struct... %v", c)

	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&request)
	if err != nil {
		util.PrintWarning("parseConnection: Error decoding the request: \n > " + err.Error())
		return request, err
	}
	d := parseData(&request.Message)
	fmt.Println("Vendor: ", d["vendor"])
	// fmt.Println("Vendor: " + string(request.Message.Data["vendor"] ))

	// Parse the encoded message

	fmt.Println(request.Stringify())
	util.PrintDebug("--------------------")
	util.PrintSuccess("[ipcserver.go] Parsed the message signature!")
	fmt.Printf("Message ID: %s\n", string(request.MessageSignature))

	// Check the message signature
	// TODO: Verify the message signature

	return request, nil
}

func parseData(msg *ipc.IPCMessage) ipc.GenericData {
	var data ipc.GenericData
	// var dataType ipc.DataType

	switch msg.Datatype {
	case ipc.DATA_TEXT:
		// Parse the integer data
		fmt.Println("Data is string")
	case ipc.DATA_INT:
		// Parse the JSON data
		fmt.Println("Data is integer")
		// data = ipc.JSONData(msg.Data)
	case ipc.DATA_JSON:
		// Parse the string data
		fmt.Println("Data is json / generic data")
		// json.Unmarshal(msg.Data, &data)

		// var temp interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println("Error unmarshaling JSON data:", err)
		} else {
			// fmt.Println("Temporary data:", temp)
			// data = temp.(map[string]interface{})
			fmt.Printf("Data: %v\n", data)
		}

	case ipc.DATA_YAML:
		// Parse the YAML data
		fmt.Println("Data is YAML / generic data")
		err := yaml.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println("Error unmarshaling YAML data:", err)
		}
	case ipc.DATA_BIN:
		// Parse the binary data
		fmt.Println("Data is binary / generic data")
	default:
		// Default to generic data
		fmt.Println("Data is generic")
		gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(&data)
	}

	if data == nil {
		fmt.Println("Data is nil")
	}

	return data
}

// Calculate the response time
func responseTime(reqTime int64) {
	currTime := util.UnixNanoTimestamp()
	diff := currTime - reqTime
	fmt.Printf("Response time: %d\n", diff)
	fmt.Printf("Seconds: %f\n", float64(diff)/1e9)
	fmt.Printf("Milliseconds: %f\n", float64(diff)/1e6)
	fmt.Printf("Microseconds: %f\n", float64(diff)/1e3)
}

// handleConnection handles the incoming connection
func handleConnection(c net.Conn) {
	defer c.Close()

	util.PrintColorf(util.LightCyan, "[🔌SOCKETS] Handling connection...")

	for {
		request, err := parseConnection(c)
		if err != nil {
			if err == io.EOF {
				util.PrintDebug("Connection closed by client")
				break
			}
			util.PrintError("Error parsing request: " + err.Error())
			break
		}

		util.PrintDebug("Request parsed: " + strconv.Itoa(request.Checksum32))

		// Process the request...
		util.PrintColorf(util.BgGreen, "Received: %+v\n", request)

		// Finally, respond to the client
		err = respond(c, request)
		if err != nil {
			util.PrintError("handleConnection: " + err.Error())
			break
		}
	}

}

func respond(c net.Conn, req ipc.IPCRequest) error {
	util.PrintDebug("Responding to the client...")
	// responseMsg := fmt.Sprintf("Acknowledged message with checksum: %d", req.Checksum32)
	// r := ipc.IPCResponse{
	// 	Request:    req,
	// 	Success:    true,
	// 	Message:    responseMsg,
	// 	Checksum32: req.Checksum32,
	// }
	var response *ipc.IPCRequest
	var err error
	if req.Checksum32 == int(crc(req.Message.Data)) {
		response, err = NewIPCMessage("bivrost", ipc.MSG_ACK, []byte("OK"))
	} else {
		fmt.Printf("Request checksum: %v\nCalculated checksum: %v\n", req.Checksum32, crc(req.Message.Data))
		response, err = NewIPCMessage("bivrost", ipc.MSG_ERROR, []byte("CHKSUM ERROR"))
	}
	if err != nil {
		return err
	}

	var responseBuffer bytes.Buffer
	encoder := gob.NewEncoder(&responseBuffer)
	err = encoder.Encode(response)
	if err != nil {
		return err
	}

	_, err = c.Write(responseBuffer.Bytes())
	if err != nil {
		return err
	}
	util.PrintColor(util.BgGreen, "🚀 Response sent!")

	responseTime(req.Timestamp)

	return nil
}

// TODO's: 1]
// - check the config for activated modules
// - check if the module is already activated
// - check if connection with the module already exists
// - create socket names based on the module name
