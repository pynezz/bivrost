/*
  IPC Package provides the IPC communication between the connector and the other modules.
*/

package ipc

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/pynezz/bivrost/internal/util"
)

const (
	AF_UNIX  = "unix"     // UNIX domain sockets
	AF_DGRAM = "unixgram" // UNIX domain datagram sockets as specified in net package

	STREAM = "SOCK_STREAM" // Stream socket 		(like TCP)
	DGRAM  = "SOCK_DGRAM"  // Datagram socket 		(like UDP)

	// Path = "/tmp/bivrost.sock"

	// Network values if applicable
	Network = "tcp"
	Address = "localhost:50052"
	Timeout = 1 * time.Second
)

const (
	// Bit flags for the socket type
	module = 0 << iota // Module
	format
	protocol
)

var Socks map[string]*UnixSocket

func init() {
	Socks = make(map[string]*UnixSocket)
}

// UnixSocket represents a UNIX domain socket server
type UnixSocket struct {
	name string // For naming the server, e.g. "Module IPC Threat Intel"
	desc string // For describing the server, e.g. "IPC server for the Threat Intel Module"

	af   string // Address family (UNIX, INET, WINSOCK)
	path string // Path to the socket file (if using UNIX domain socket)

	// addr     net.UnixAddr     // UNIX domain socket address
	// conn     net.UnixConn     // UNIX domain socket connection
	// listener net.UnixListener // UNIX domain socket listener
	// syscall  int              // System call

	// Run uintptr // Run the server

	IsConnected bool      // Is the server connected
	connection  *net.Conn // Connection to the server
}

// Will establish a connection to the socket at the given path
// with the network protocol "unix" as specified in the net package
// func (s *UnixSocket) ConnectSocket() (*net.Conn, error) {
// 	// Connect to the socket
// 	conn, err := net.Dial(AF_UNIX, path); if err != nil {
// 		return nil, err
// 	}

// 	defer conn.Close()

// 	fmt.Fprintf(conn, "Hello\n")

// 	return &conn, nil
// }

func (s *UnixSocket) GetConn() *net.Conn {
	return s.connection
}

func (s *UnixSocket) Listen() error {
	l, err := net.Listen(s.af, s.path)
	if err != nil {
		return err
	}
	defer l.Close()

	fmt.Println("Listening on ", s.path)

	for {
		fmt.Println("Waiting for connection...")
		c, err := l.Accept()
		if err != nil {
			log.Fatalf("failed to accept: %v", err)
		}

		fmt.Println("Incoming connection: ", c.LocalAddr().String())
		go s.handleListener(c)
	}
}

// NewSocket creates a new UNIX domain socket server
func NewSocket(name string, desc string) (*UnixSocket, error) {
	tmpDir := "/tmp/bivrost"
	err := os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	fmt.Println("Temp dir: ", tmpDir)

	// path := path.Join(tmpDir, "bivrost-"+name+".sock")
	path := path.Join(tmpDir, "bivrost.sock")

	fmt.Println("Full path: ", path)

	var choice string // For the choice of the socket type
	_, err = fmt.Sscanf("Do you want to continue? [Y/n] > ", "%s", &choice)
	if err != nil {
		return nil, err
	}
	if choice[0] == 'n' {
		return nil, fmt.Errorf("[aborted] user chose to exit")
	}

	s := &UnixSocket{
		name:       name,
		desc:       desc,
		af:         AF_UNIX,
		path:       path,
		connection: nil,
	}

	return s, nil
}

// StartUDSServer starts the UDS server
func (s *UnixSocket) Initialize() error {
	util.PrintInfo("Starting UNIX Domain Sockets SERVER...")

	l, err := net.Listen(s.af, s.path)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalf("failed to accept: %v", err)
		}
		go s.handleConnection(c)
	}
}

func (s *UnixSocket) handleListener(c net.Conn) {
	log.Printf("Received connection from %v", c.LocalAddr())
	fmt.Fprintf(c, "Hello, %s\n", c.RemoteAddr())
	c.Close()
}

// handleConnection handles the connection
func (s *UnixSocket) handleConnection(c net.Conn) {
	log.Printf("Received connection from %v", c.RemoteAddr())
	fmt.Fprintf(c, "Hello, %s\n", c.RemoteAddr())
	c.Close()
}

func (s *UnixSocket) Cleanup() {

	// Remove the socket file
	os.RemoveAll(s.path)
}
