package gowebsocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ApparentlyAndy/go-websocket/internal"
)

type WSConnection *internal.Websocket

func WSUpgrade(w http.ResponseWriter, r *http.Request, onConnect func(WSConnection), onReceive func(interface{})) {
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
			onReceive(data)
		}
	}
}
