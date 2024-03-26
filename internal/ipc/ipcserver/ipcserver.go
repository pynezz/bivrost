package ipcserver

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
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

/* TYPES
 * Types for the IPC communication between the connector and the other modules.
 */
type IPCServer struct {
	path string

	conn net.Listener
}

var IPCServerBuffer = make(map[string]*IPCServer)

func NewIPCServer(path string) *IPCServer {
	return &IPCServer{
		path: path,
	}
}

// will this remove the socket file?
// Where does it write it?
// A:
func (s *IPCServer) InitServerSocket() bool {
	// Making sure the socket is clean before starting
	if err := os.RemoveAll(s.path); err != nil {
		util.PrintError("InitServerSocket(): Failed to remove old socket: " + err.Error())
		return false
	}

	IPCServerBuffer[s.path] = s

	return true
}

func (s *IPCServer) Listen() {
	s.conn, _ = net.Listen(AF_UNIX, s.path)
	util.PrintColorf(util.Gray, "[🔌SOCKETS] Starting listener on %s", s.path)

	for {
		conn, err := s.conn.Accept()
		util.PrintColorf(util.Gray, "[🔌SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			util.PrintError("Listen(): " + err.Error())
			continue
		}

		go handleConnection(conn)
		util.PrintColorBold(util.DarkGreen, "🎉 IPC server started!")
	}
}

func Cleanup() {
	for _, server := range IPCServerBuffer {
		util.PrintItalic("Cleaning up IPC server: " + server.path)
		err := os.Remove(server.path)
		if err != nil {
			util.PrintError("Cleanup(): " + err.Error())
		}

		util.PrintItalic("Closing connection: " + server.conn.Addr().String())
		server.conn.Close()
	}
	IPCServerBuffer = make(map[string]*IPCServer)
	util.PrintItalic("IPC server cleanup complete")
}

type ipcRequest struct {
	Header  ipcHeader  // The header - containing type and identifier
	Message ipcMessage // The message
	Chksum  []byte     // Checksum of the message
}

type ipcHeader struct {
	Identifier  [4]byte // Identifier of the module
	MessageType byte    // Type of the message
}

type ipcMessage struct {
	Data       []byte // Data of the message
	StringData string // String representation of the data if applicable
}

// Function to create a new IPCMessage based on the identifier key
func NewIPCMessage(identifierKey string, messageType byte, data []byte) (*ipcRequest, error) {
	identifier, ok := IDENTIFIERS[identifierKey]
	if !ok {
		return nil, fmt.Errorf("invalid identifier key: %s", identifierKey)
	}

	var id [4]byte
	copy(id[:], identifier[:4]) // Ensure no out of bounds panic

	message := &ipcMessage{
		Data:       data,
		StringData: string(data),
	}

	// TODO:
	// - Define the size
	// - Define the algorithm
	checksum := []byte{0x00, 0x00, 0x00, 0x00}

	return &ipcRequest{
		Header: ipcHeader{
			Identifier:  id,
			MessageType: messageType,
		},
		Message: *message,
		Chksum:  checksum,
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
	defer c.Close() // May not be needed due to Close() in Cleanup()
	reader := bufio.NewReader(c)
	for {
		// Example of reading a fixed-size header first
		header := make([]byte, 5) // Adjust size based on your protocol
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read header error: %v", err)
			} else {
				log.Println("Client disconnected")
			}
			break // Exit the loop on error or EOF
		}

		// Parse the header (example function, you'll need to implement this)
		request, err := parseConnection(header)
		if err != nil {
			log.Printf("Parse error: %v", err)
			break
		}

		// If your protocol specifies the data length in the header, read that many bytes next
		// Here's a placeholder for reading the rest of the data based on your protocol
		data := make([]byte, len(request.Message.Data)) // Define dataLength based on your header
		_, err = io.ReadFull(reader, data)
		if err != nil {
			log.Printf("Read data error: %v", err)
			break
		}

		// Process the request...
		log.Printf("Received: %+v", request)

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

// TODO's:
// - check the config for activated modules
// - check if the module is already activated
// - check if connection with the module already exists
// - create socket names based on the module name
