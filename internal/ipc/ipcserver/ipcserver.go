package ipcserver

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net"
	"os"

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

type ipcRequest struct {
	Header     ipcHeader  `` // The header - containing type and identifier
	Message    ipcMessage // The message
	Checksum32 int        // Checksum of the message byte data
}

type ipcHeader struct {
	Identifier  [4]byte // Identifier of the module
	MessageType byte    // Type of the message
}

type ipcMessage struct {
	Data       []byte // Data of the message
	StringData string // String representation of the data if applicable
}

func init() {
	// IDENTIFIERS = map[string][]byte{
	// 	"threat_intel": THREAT_INTEL,
	// }

	// MSGTYPE = map[string]byte{
	// 	"request":  0x01,
	// 	"response": 0x02,
	// }
	connections = Connections{}                       // (i) <- likely not needed
	connections.sockets = make(map[string]*IPCServer) // (i) <- likely not needed
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
	util.PrintColorf(util.DarkBlue, "[🔌SOCKETS] Starting listener on %s", s.path)

	for {
		// (i) This may be where the garbage collector tries to free the memory, which is already freed manually. (ref: todo 2)
		// (i) Meaning that we might not need to do it manually. We should anyways still delete the files though. So we need a ref to that.
		conn, err := s.conn.Accept()
		util.PrintColorf(util.DarkBlue, "[🔌SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			util.PrintError("Listen(): " + err.Error())
			continue
		}

		go handleConnection(conn)
		util.PrintColorBold(util.DarkGreen, "🎉 IPC server started!")
	}
}

func Cleanup() {
	for _, server := range connections.sockets {
		util.PrintItalic("Cleaning up IPC server: " + server.path)
		err := os.Remove(server.path)
		if err != nil {
			util.PrintError("Cleanup(): " + err.Error())
		}

		util.PrintItalic("Closing connection: " + server.conn.Addr().String())
		server.conn.Close()
	}
	util.PrintItalic("IPC server cleanup complete")
}

func (r *ipcRequest) Stringify() string {
	h := fmt.Sprintf("HEADER:\n\tIdentifier: %v\nMessageType: %v\n", r.Header.Identifier, r.Header.MessageType)
	m := fmt.Sprintf("MESSAGE:\n\tData: %v\n\tStringData: %v\n", r.Message.Data, r.Message.StringData)
	c := fmt.Sprintf("CHECKSUM: %v\n", r.Checksum32)
	return h + m + c
}

func (m *ipcMessage) Stringify() func() string {
	f := func() string {
		return fmt.Sprintf("Data: %v", m.Data)
	}

	return f
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
func NewIPCMessage(identifierKey string, messageType byte, data []byte) (*ipcRequest, error) {
	identifier, ok := IDENTIFIERS[identifierKey]
	if !ok {
		return nil, fmt.Errorf("invalid identifier key: %s", identifierKey)
	}

	var id [4]byte
	copy(id[:], identifier[:4]) // Ensure no out of bounds panic

	crcsum32 := crc(data)

	message := &ipcMessage{
		Data:       data,
		StringData: string(data),
	}

	return &ipcRequest{
		Header: ipcHeader{
			Identifier:  id,
			MessageType: messageType,
		},
		Message:    *message,
		Checksum32: crcsum32,
	}, nil
}

func parseConnection(message []byte) (ipcRequest, error) {
	var request ipcRequest
	var err error

	if len(message) < 5 {
		fmt.Println("parseConnection: Message too short")
		return request, nil
	}

	headerBuf := bytes.NewReader(message[:5])

	err = binary.Read(headerBuf, binary.BigEndian, &request.Header)
	if err != nil {
		return request, err
	}

	request.Message.Data = message[5:]

	return request, nil
}

func handleConnection(c net.Conn) {
	// defer c.Close() // May not be needed due to Close() in Cleanup()
	util.PrintColorf(util.DarkBlue, "[🔌SOCKETS] Handling connection...")
	reader := bufio.NewReader(c)
	for {
		// Example of reading a fixed-size header first
		header := make([]byte, 5) // Adjust size based on your protocol
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read header error: %v\n", err)
			} else {
				log.Println("Client disconnected")
			}
			break // Exit the loop on error or EOF
		}

		// Parse the header (example function, you'll need to implement this)
		request, err := parseConnection(header)
		if err != nil {
			log.Printf("Parse error: %v\n", err)
			break
		}

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
}

type module struct {
	name string
	desc string
}

var (
	THREAT_INTEL = []byte{0x74, 0x68, 0x72, 0x69} // "thri"
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
