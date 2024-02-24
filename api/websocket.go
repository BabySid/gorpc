package api

type WSMessageType int

const (
	WSTextMessage   WSMessageType = 1
	WSBinaryMessage WSMessageType = 2
)

type WSMessage struct {
	Type WSMessageType
	Data []byte
}
