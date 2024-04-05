package ipc

import (
	"fmt"
)

type IPCRequest struct {
	MessageSignature []byte     // The message signature, used to declare an ipcRequest
	Header           IPCHeader  // The header - containing type and identifier
	Message          IPCMessage // The message
	Timestamp        int64      // Timestamp of the message
	Checksum32       int        // Checksum of the message byte data
}

type IPCHeader struct {
	Identifier  [4]byte // Identifier of the module - available from the IPCClient for qol purposes
	MessageType byte    // Type of the message
}

type IPCMessage struct {
	Datatype   DataType // Type of the data ("json", "string", "int", etc.)
	Data       []byte   // The actual data
	StringData string   // String representation of the data if applicable
}

type IPCResponse struct {
	Request    IPCRequest // The request that was sent
	Success    bool       // Was the request successful
	Message    string     // Message from the server
	Checksum32 int        // Checksum of the message byte data
}

// GenericData is a generic map for data. It can be used to store any data type.
//
// Ex:
//
//	var data GenericData = map[string]interface{}{"someKey": "someValue"}
//	fmt.Println("Received data:", data["someKey"]) // Will print someValue
type GenericData map[string]interface{}

type MsgType int
type DataType int

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

const (
	DATA_TEXT = 0x01 // Text data
	DATA_INT  = 0x02 // Integer data
	DATA_JSON = 0x03 // JSON data	(used for structured data)
	DATA_YAML = 0x04 // YAML data	(used for configuration files)
	DATA_BIN  = 0x05 // Binary data	(such as images, files, etc.)
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
	// m := "MESSAGE:\n"
	// for i, data := range r.Message.Data {
	// 	m += fmt.Sprintf("\tData[%d]: %v\n", i, data)
	// }
	m := fmt.Sprintf("MESSAGE:\n\tData: %v\n\tStringData: %v\n", r.Message.Data, r.Message.StringData)
	c := fmt.Sprintf("CHECKSUM: %v\n", r.Checksum32)
	return h + m + c
}
