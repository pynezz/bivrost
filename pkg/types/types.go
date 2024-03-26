package types

type IPCRequest struct {
	Header IPCHeader
	Data   []byte
	Error  error
}

type IPCResponse struct {
	Header IPCHeader
	Data   []byte
	Error  error
}

// 5 bytes
type IPCHeader struct {
	Identifier  [4]byte
	MessageType byte
}
