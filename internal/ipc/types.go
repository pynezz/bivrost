package ipc

type IPCRequest struct {
	Header     IPCHeader  // The header - containing type and identifier
	Message    IPCMessage // The message
	Checksum32 int        // Checksum of the message byte data
}

type IPCHeader struct {
	Identifier  [4]byte // Identifier of the module - available from the IPCClient for qol purposes
	MessageType byte    // Type of the message
}

type IPCMessage struct {
	Data       []byte // Data of the message
	StringData string // String representation of the data if applicable
}

type MsgType int

const (
	MSG_CONN    = 0x01 // Connection message
	MSG_ACK     = 0x02 // Acknowledgement message
	MSG_CONNACK = 0x03 // Connection acknowledgement message
	MSG_MSG     = 0x04 // Message
	MSG_MSGACK  = 0x05 // Message acknowledgement

	MSG_PING = 0x08 // Ping message
	MSG_PONG = 0x09 // Pong message

	MSG_DISCONNECT = 0xD1 // Disconnect message

	// Error message
	// The sender is obviously still connected, but something went wrong and the message was not handled
	MSG_ERROR = 0xEE

	MSG_UNKNOWN = 0xFF // Unknown message - for signifying unknown type, maybe an error, but the receiver will try to wing it
)

var MSGTYPE = map[string]byte{
	"conn":       byte(MSG_CONN),
	"ack":        byte(MSG_ACK),
	"connack":    byte(MSG_CONNACK),
	"msg":        byte(MSG_MSG),
	"msgack":     byte(MSG_MSGACK),
	"ping":       byte(MSG_PING),
	"pong":       byte(MSG_PONG),
	"disconnect": byte(MSG_DISCONNECT),
	"error":      byte(MSG_ERROR),
	"unknown":    byte(MSG_UNKNOWN),
}
