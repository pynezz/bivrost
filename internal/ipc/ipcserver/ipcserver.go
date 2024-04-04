package ipcserver

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/pynezz/bivrost/internal/ipc"
	"github.com/pynezz/bivrost/internal/util"
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

	s.conn, _ = net.Listen(AF_UNIX, s.path)
	util.PrintColorf(util.LightCyan, "[🔌SOCKETS] Starting listener on %s", s.path)

	for {
		util.PrintDebug("Waiting for connection...")
		// (i) This may be where the garbage collector tries to free the memory, which is already freed manually. (ref: todo 2)
		// (i) Meaning that we might not need to do it manually. We should anyways still delete the files though. So we need a ref to that.
		conn, err := s.conn.Accept()
		util.PrintColorf(util.LightCyan, "[🔌SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			util.PrintError("Listen(): " + err.Error())
			continue
		}

		go handleConnection(conn)
		util.PrintColorBold(util.DarkGreen, "🎉 IPC server running!")
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

func crc(b []byte) int {
	c := crc32.NewIEEE()
	s, err := c.Write(b)
	if err != nil {
		util.PrintError("crc(): " + err.Error())
		return 0xFF // Error
	}
	return s
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
		Checksum32: crcsum32,
	}, nil
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

	fmt.Println(request.Stringify())
	util.PrintDebug("--------------------")
	util.PrintSuccess("[ipcserver.go] Parsed the message signature!")
	fmt.Printf("Message ID: %s\n", string(request.MessageSignature))

	// Check the message signature
	// TODO: Verify the message signature

	return request, nil
}

// handleConnection handles the incoming connection
func handleConnection(c net.Conn) {
	// defer c.Close() // May not be needed due to Close() in Cleanup()
	util.PrintColorf(util.LightCyan, "[🔌SOCKETS] Handling connection...")

	request, err := parseConnection(c)
	if err != nil {
		log.Printf("Parse error: %v\n", err)
	}

	util.PrintDebug("Request parsed: " + strconv.Itoa(request.Checksum32))

	// Process the request...
	util.PrintColorf(util.BgGreen, "Received: %+v\n", request)

	// Finally, respond to the client
	err = respond(c)
	if err != nil {
		util.PrintError("handleConnection: " + err.Error())
	}
}

func respond(c net.Conn) error {
	util.PrintDebug("Responding to the client...")
	var responseBuffer bytes.Buffer
	response, err := NewIPCMessage("bivrost", ipc.MSG_ACK, []byte("Message received"))
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(&responseBuffer)
	err = encoder.Encode(response)
	if err != nil {
		return err
	}

	_, err = c.Write(responseBuffer.Bytes())
	if err != nil {
		return err
	}

	util.PrintColor(util.BgGreen, "Response sent!")

	return nil

}

// TODO's: 1]
// - check the config for activated modules
// - check if the module is already activated
// - check if connection with the module already exists
// - create socket names based on the module name
