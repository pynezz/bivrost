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

// UDSServer represents a UNIX domain socket server
type UDSServer struct {
	name string // For naming the server, e.g. "Module IPC Threat Intel"
	desc string // For describing the server, e.g. "IPC server for the Threat Intel Module"

	af   string // Address family (UNIX, INET, WINSOCK)
	path string // Path to the socket file (if using UNIX domain socket)

	// addr     net.UnixAddr     // UNIX domain socket address
	// conn     net.UnixConn     // UNIX domain socket connection
	// listener net.UnixListener // UNIX domain socket listener
	// syscall  int              // System call

	// Run uintptr // Run the server
}

func NewSocket(name string, desc string) (*UDSServer, error) {
	tmpDir, err := os.MkdirTemp("/tmp", "bivrost_ipc_sock_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("Temp dir: ", tmpDir)

	path := path.Join(tmpDir, "bivrost-"+name+".sock")

	fmt.Println("Full path: ", path)

	s := &UDSServer{
		name: name,
		desc: desc,
		af:   AF_UNIX,
		path: path,
	}

	return s, nil
}

// StartUDSServer starts the UDS server
func (s *UDSServer) Initialize() {
	util.PrintInfo("Starting UNIX Domain Sockets server...")
	l, err := net.Listen(s.af, s.path); if err != nil {
		log.Fatalf("failed to listen: %v", err)
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

// handleConnection handles the connection
func (s *UDSServer) handleConnection(c net.Conn) {
	log.Printf("Received connection from %v", c.RemoteAddr())
	fmt.Fprintf(c, "Hello, %s\n", c.RemoteAddr())
	c.Close()
}

func (s *UDSServer) Cleanup() {
	// Cleanup the server
}

func (s *UDSServer) IsNil() bool {
	return s == nil
}
