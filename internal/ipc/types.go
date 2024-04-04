package ipc

import (
	"fmt"
)

type IPCRequest struct {
	MessageSignature []byte     // The message signature, used to declare an ipcRequest
	Header           IPCHeader  // The header - containing type and identifier
	Message          IPCMessage // The message
	Checksum32       int        // Checksum of the message byte data
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

var IDENTIFIERS = map[string][4]byte{
	"threat_intel": [4]byte([]byte("THRI")), // Threat Intel (should equal to 0x54, 0x48, 0x52, 0x49)
}

func (r *IPCRequest) Stringify() string {
	h := fmt.Sprintf("HEADER:\n\tIdentifier: %v\nMessageType: %v\n", r.Header.Identifier, r.Header.MessageType)
	m := fmt.Sprintf("MESSAGE:\n\tData: %v\n\tStringData: %v\n", r.Message.Data, r.Message.StringData)
	c := fmt.Sprintf("CHECKSUM: %v\n", r.Checksum32)
	return h + m + c
}
