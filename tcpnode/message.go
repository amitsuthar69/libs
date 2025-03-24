package tcpnode

import "strings"

type MessageType int

const (
	HandshakeMessage MessageType = iota
	DataMessage
	ClientMessage
	PingMessage
)

type Message struct {
	Payload  []byte
	Type     MessageType
	SenderId string
}

// ParseMessage returns the bytes as Handshake message, PingMessage or a Datamessage.
//
// It checks for the "HS:" prefix in first three bytes.
//
// format- HS:<nodeId>:<serverAddr>
func ParseMessage(data []byte) Message {
	if len(data) >= 3 && string(data[:3]) == "HS:" {
		parts := strings.Split(string(data[3:]), ":")
		if len(parts) == 2 {
			return Message{
				Payload:  []byte(parts[1]),
				Type:     HandshakeMessage,
				SenderId: parts[0],
			}
		}
	} else if string(data[:4]) == "PING" {
		return Message{Payload: data, Type: PingMessage}
	}

	return Message{Payload: data, Type: DataMessage}
}
