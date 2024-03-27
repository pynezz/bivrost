package ipcclient

import "net"

const (
	AF_UNIX  = "unix"     // UNIX domain sockets
	AF_DGRAM = "unixgram" // UNIX domain datagram sockets as specified in net package
)

type IPCClient struct {
	Name string // Name of the module
	Desc string // Description of the module

	Identifier [4]byte // Identifier of the module

	conn net.Conn // Connection to the IPC server (UNIX domain socket)
}
