package ipcclient

import (
	"fmt"
	"net"
	"os"
	"path"

	"github.com/pynezz/bivrost/internal/util"
)

const (
	AF_UNIX  = "unix"     // UNIX domain sockets
	AF_DGRAM = "unixgram" // UNIX domain datagram sockets as specified in net package
)

type IPCClient struct {
	Name string
	Desc string

	conn net.Conn
}

var clientsBuffer = make(map[string]*IPCClient)

func NewIPCClient() *IPCClient {
	return &IPCClient{}
}

// We'll get the path from the config, but for now let's just hard code it
func (c *IPCClient) Connect(module string) error {
	tmpDir := os.TempDir()
	bivrostTmpDir := path.Join(tmpDir, "bivrost")
	socketPath := path.Join(bivrostTmpDir, "bivrost.sock")

	// Add the connection to the client buffer

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("Dial error:", err)
		return err
	}
	c.conn = conn

	fmt.Printf("Connected to %s\n", socketPath)
	clientsBuffer[module] = c
	return nil
}

// sendMessage sends a string message to the connection
func (c *IPCClient) SendMessage(message string) error {
	if c.conn == nil {
		fmt.Println("IPCClient not connected\nDid you forget to call Connect()?")
		// Prompt the user if they want to retry

		var retry string
		util.PrintWarning("Do you want to retry? [Y/n]")
		fmt.Scanln(&retry)
		if retry[0] == 'n' {
			return fmt.Errorf("IPCClient not connected")
		}
		c.Connect(message)
	}
	_, err := c.conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Write error:", err)
		return err
	}
	fmt.Println("Message sent:", message)

	return nil
}

// Close the connection
func (c *IPCClient) Close() {
	c.conn.Close()
}

// Close all connections
func CloseAll() {
	for _, client := range clientsBuffer {
		client.Close()
	}
}
