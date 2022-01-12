package internal

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"strings"
)

var bufferSize = 4096

type Connection interface {
	Close() error
}

type Websocket struct {
	Conn   Connection
	bufrw  *bufio.ReadWriter
	header http.Header
	status uint16
}

func getWSAcceptHash(key string) string {
	wsMagicString := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	sha1 := sha1.Sum([]byte(key + wsMagicString))
	return base64.StdEncoding.EncodeToString(sha1[:])
}

func HijackConnection(w http.ResponseWriter, r *http.Request) (*Websocket, error) {
	hijack, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("doesn't support hijacking")
	}

	conn, bufrw, err := hijack.Hijack()

	if err != nil {
		return nil, err
	}

	return &Websocket{Conn: conn, bufrw: bufrw, header: r.Header, status: http.StatusSwitchingProtocols}, nil
}

func (ws *Websocket) Handshake() error {
	hash := getWSAcceptHash(ws.header.Get("Sec-WebSocket-Key"))
	headers := []string{
		"HTTP/1.1 101 Web Socket Protocol Handshake",
		"Server: go/websocketServer",
		"Upgrade: Websocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Accept: " + hash,
		"",
		"",
	}
	return ws.write([]byte(strings.Join(headers, "\r\n")))
}

func (ws *Websocket) write(data []byte) error {
	if _, err := ws.bufrw.Write(data); err != nil {
		return err
	}
	return ws.bufrw.Flush()
}

func (ws *Websocket) read(size int) ([]byte, error) {
	data := make([]byte, 0)

	for {
		if len(data) == size {
			break
		}

		sz := bufferSize
		remaining := size - len(data)

		if sz > remaining {
			sz = remaining
		}

		temp := make([]byte, sz)

		n, err := ws.bufrw.Read(temp)

		if err != nil && err != io.EOF {
			return data, err
		}

		data = append(data, temp[:n]...)
	}

	return data, nil
}

func (ws *Websocket) Recv() (Frame, error) {
	frame := Frame{}

	head, err := ws.read(2)
	if err != nil {
		return frame, err
	}

	frame.IsFragment = (head[0] & 0x80) == 0x00
	frame.OpCode = (head[0] & 0x0F)
	frame.Reserved = (head[0] & 0x70)
	frame.IsMasked = (head[1] & 0x80) == 0x80

	length := uint64(head[1] & 0x7F)

	if length == 126 {
		data, err := ws.read(2)
		if err != nil {
			return frame, err
		}
		length = uint64(binary.BigEndian.Uint16(data))
	}

	if length == 127 {
		data, err := ws.read(8)
		if err != nil {
			return frame, err
		}
		length = uint64(binary.BigEndian.Uint64(data))
	}

	maskingKey, err := ws.read(4)

	if err != nil {
		return frame, err
	}

	frame.Length = int(length)

	payload, err := ws.read(frame.Length)

	if err != nil {
		return frame, err
	}

	for i := 0; i < int(length); i++ {
		payload[i] ^= maskingKey[i%4]
	}

	frame.Payload = payload

	return frame, err
}

func (ws *Websocket) Send(data []byte, binary bool) {
	frame := Frame{}

	frame.IsFragment = false
	frame.OpCode = 0x1

	if binary {
		frame.OpCode = 0x2
	}

	frame.Length = len(data)
	frame.Payload = data

	dataFrame := frame.makeDataFrame()
	ws.write(dataFrame)
}
