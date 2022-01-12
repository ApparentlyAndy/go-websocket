package gowebsocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ApparentlyAndy/go-websocket/internal"
)

type Message struct {
	Type string
	Data interface{}
}

func WSUpgrade(w http.ResponseWriter, r *http.Request, messageEvent func(Message)) {
	ws, err := internal.HijackConnection(w, r)
	ws.Handshake()

	message := Message{}

	if err != nil {
		log.Println(err)
	}

	defer ws.Conn.Close()

	for {
		frame, err := ws.Recv()
		if err != nil {
			fmt.Printf("%s", err)
		}

		data, err := frame.ReadData()

		if err != nil {
			log.Println(err)
		} else {
			if frame.OpCode == 0x1 {
				message.Type = "string"
				message.Data = string(data.([]byte))
			} else if frame.OpCode == 0x2 {
				message.Type = "binary"
				message.Data = data.([]byte)
			} else {
				message.Type = "unknown"
			}

			messageEvent(message)
		}
	}
}
