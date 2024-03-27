package ipcclient

import (
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

// type ipcRequest ipc.IPCRequest
// type ipcHeader ipc.IPCHeader
// type ipcMessage ipc.IPCMessage

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

// We'll get the path from the config, but for now let's just hard code it
func (c *IPCClient) Connect(module string) error {
	tmpDir := os.TempDir()
	bivrostTmpDir := path.Join(tmpDir, "bivrost")
	socketPath := path.Join(bivrostTmpDir, "bivrost.sock")

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

	fmt.Printf("Connected to %s\n", socketPath)

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
	var retry string
	util.PrintWarning("Do you want to retry? [Y/n]")
	fmt.Scanln(&retry)
	if retry[0] == 'n' {
		return false
	}
	return true
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

/*
NewMessage creates a new IPC message

message string: The message to send

t: The message type
*/
func (c *IPCClient) CreateReq(message string, t ipc.MsgType) *ipc.IPCRequest {
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
	}

	// return &ipcRequest{
	// 	ipc.Header: ipcHeader{
	// 		ipc.Identifier:  c.Identifier,
	// 		ipc.MessageType: byte(t),
	// 	},
	// 	ipc.Message: ipcMessage{
	// 		ipc.Data: []byte(message),
	// 	},
	// }
}

// Close the connection
func (c *IPCClient) Close() {
	c.conn.Close()
}
