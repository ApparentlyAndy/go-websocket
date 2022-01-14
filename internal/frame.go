package internal

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

type Frame struct {
	IsFragment bool
	OpCode     byte
	Reserved   byte
	IsMasked   bool
	Length     int
	Payload    []byte
}

type Message struct {
	Type    string
	Payload []byte
}

func (f *Frame) ReadData() (*Message, error) {
	message := Message{}
	if f.OpCode == 0x1 {
		message.Type = "string"
	} else if f.OpCode == 0x2 {
		message.Type = "binary"
	} else {
		return &message, errors.New("unable to read data.")
	}
	message.Payload = f.Payload
	return &message, nil
}

func (f *Frame) makeDataFrame() []byte {
	data := make([]byte, 2)
	data[0] = 0x80 ^ f.OpCode

	if f.Length <= 125 {
		data[1] = byte(f.Length)
		data = append(data, f.Payload...)
		// Payload length more than 125 and less than 65536
	} else if f.Length > 125 && f.Length < 65536 {
		data[1] = byte(126)
		payloadLen := make([]byte, 2)
		binary.BigEndian.PutUint16(payloadLen, uint16(f.Length))
		data = append(data, payloadLen...)
		data = append(data, f.Payload...)
	} else if f.Length >= 65536 {
		data[1] = byte(127)
		payloadLen := make([]byte, 8)
		binary.BigEndian.PutUint64(payloadLen, uint64(f.Length))
		data = append(data, payloadLen...)
		data = append(data, f.Payload...)
	}

	return data
}
