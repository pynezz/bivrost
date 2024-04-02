package ipcserver

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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

var MSGTYPE map[string]byte

// var IPCServerBuffer = make(map[string]*IPCServer)

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

	// IPCServerBuffer[s.path] = s //! // TODO: segmentation violation on cleanup
	// connections.sockets[s.path] = s // (i) <- likely not needed

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
	util.PrintItalic("IPC server cleanup complete")
}

// func (r *ipc.IPCRequest) Stringify() string {
// 	h := fmt.Sprintf("HEADER:\n\tIdentifier: %v\nMessageType: %v\n", r.Header.Identifier, r.Header.MessageType)
// 	m := fmt.Sprintf("MESSAGE:\n\tData: %v\n\tStringData: %v\n", r.Message.Data, r.Message.StringData)
// 	c := fmt.Sprintf("CHECKSUM: %v\n", r.Checksum32)
// 	return h + m + c
// }

// func (m *ipc.IPCMessage) Stringify() func() string {
// 	f := func() string {
// 		return fmt.Sprintf("Data: %v", m.Data)
// 	}

// 	return f
// }

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

	util.PrintDebug("Trying to encode the bytes to a request struct...")
	request.Stringify()
	util.PrintDebug("--------------------")

	//

	// util.PrintColorf(util.LightCyan, "[🔌SOCKETS] Got a message of %d length", len(message))
	// msgId, err := parseMessageSignature(message)
	if err != nil {
		return request, err
	}
	util.PrintSuccess("[ipcserver.go] Parsed the message signature!")
	fmt.Printf("Message ID: %v\n", request.MessageSignature)

	// sigLen := len(ipc.IPCID)
	// headerBuf := bytes.NewReader(message[sigLen : 5+sigLen]) // Read the header

	// err = binary.Read(headerBuf, binary.BigEndian, &request.Header)
	// if err != nil {
	// 	return request, err
	// }

	// request.Message.Data = message[5+sigLen:] // Read the data
	// util.PrintWarning("Attempting to parse the message data 'ipcMessage'...")

	return request, nil
}

// handleConnection handles the incoming connection
func handleConnection(c net.Conn) {
	// defer c.Close() // May not be needed due to Close() in Cleanup()
	util.PrintColorf(util.LightCyan, "[🔌SOCKETS] Handling connection...")
	reader := bufio.NewReader(c) //! Might be an issue here with the underlying reader reading more than just the struct sizes
	//! // TODO: Test with just the c connection
	x := make(chan os.Signal, 1)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)

	for {
		request, err := parseConnection(c)
		if err != nil {
			log.Printf("Parse error: %v\n", err)
			break
		}

		// Example of reading a fixed-size header first
		msgSig := make([]byte, 6)            // Adjust size based on your protocol
		_, err = io.ReadFull(reader, msgSig) // This will fail.
		if err != nil {
			if err != io.EOF {
				log.Printf("Read header error: %v\n", err)
			} else {
				log.Println("Client disconnected")
			}
			break // Exit the loop on error or EOF
		}

		// validateMessageSignature(msgSig) // Validate the message signature

		// If your protocol specifies the data length in the header, read that many bytes next
		// Here's a placeholder for reading the rest of the data based on your protocol
		data := make([]byte, len(request.Message.Data)) // Define dataLength based on your header
		_, err = io.ReadFull(reader, data)
		if err != nil {
			log.Printf("Read data error: %v\n", err)
			break
		}

		// Process the request...
		log.Printf("Received: %+v\n", request)

		// Example response
		c.Write([]byte("ACK\n"))
	}
	<-x
}

type module struct {
	name string
	desc string
}

var (
	THREAT_INTEL = []byte{0x54, 0x48, 0x52, 0x49} // "THRI"
)

// func parseRequest(r *bufio.Reader) {
// 	var request ipcRequest
// 	var response ipcResponse
// 	var err error

// 	// Read the header
// 	request, err = r.ReadBytes('\n')

// 	if err != nil {
// 		response.Error = err
// 		writeResponse(c, response)
// 		return
// 	}

// 	// Read the data
// 	request.Data, err = readData(r)
// 	if err != nil {
// 		response.Error = err
// 		writeResponse(c, response)
// 		return
// 	}

// 	response.Header = request.Header
// 	response.Data = request.Data
// 	response.Error = nil

// 	writeResponse(c, response)

// }

// TODO's: 1]
// - check the config for activated modules
// - check if the module is already activated
// - check if connection with the module already exists
// - create socket names based on the module name

// TODO: 2] Fix segmentation violation on cleanup
/*
Done cleaning up. Exiting...
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x20 pc=0x783c8c]

goroutine 22 [running]:
github.com/pynezz/bivrost/internal/ipc/ipcserver.(*IPCServer).Listen(0xc0000b60a0)
        .../bivrost/internal/ipc/ipcserver/ipcserver.go:87 +0xcc
created by main.testUDS in goroutine 3
        .../bivrost/main.go:168 +0x216
*/
