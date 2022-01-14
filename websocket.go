package gowebsocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ApparentlyAndy/go-websocket/internal"
)

func WSUpgrade(w http.ResponseWriter, r *http.Request, onConnect func(*internal.Websocket), onReceive func(interface{})) {
	ws, err := internal.HijackConnection(w, r)
	ws.Handshake()

	if err != nil {
		log.Println(err)
	}

	onConnect(ws)

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
				onReceive(string(data.([]byte)))
			} else {
				onReceive(data.([]byte))
			}
		}
	}
}
