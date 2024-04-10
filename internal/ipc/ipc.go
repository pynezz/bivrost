/*
  IPC Package provides the IPC communication between the connector and the other modules.
  * THIS WHOLE FILE ALREADY NEEDS REFACTORING
*/

package ipc

import (
	"encoding/gob"
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

	// Network values if applicable
	Network = "tcp"
	Address = "localhost:50052"
	Timeout = 1 * time.Second
)

var IPCID []byte // Identifier of the IPC communication

type IPCMessageId []byte // Identifier of the message

func SetIPCID(id []byte) {
	if IPCID == nil {
		IPCID = id
		util.PrintSuccess("Set IPC ID to " + string(IPCID))
	} else {
		util.PrintWarning("IPC ID already set to " + string(IPCID))
	}
}

func GetIPCStrID() string {
	return string(IPCID)
}

func DefaultSock(name string) string {
	tmpDir := os.TempDir()                     // Temporary directory (eg. /tmp)
	subTmpDir := path.Join(tmpDir, name)       // Subdirectory in the temporary directory (eg. /tmp/<subTmpDir>)
	sock := path.Join(subTmpDir, name+".sock") // Socket file path (eg. /tmp/<subTmpDir>/<name>)
	sock = path.Clean(sock)                    // Clean the path

	return sock
}

func init() {

	gob.Register(IPCRequest{})
	gob.Register(IPCMessage{})
	gob.Register(IPCHeader{})
	gob.Register(IPCMessageId{})
	gob.Register(IPCResponse{})
}
