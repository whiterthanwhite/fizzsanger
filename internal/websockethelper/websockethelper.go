package websockethelper

import (
	"log"

	"github.com/gorilla/websocket"
)

func GetUserMessage(conn *websocket.Conn, userMsg chan []byte) {
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err.Error())
			return
		}
		log.Println("\nMessage type: ", msgType, "\nmessage: ", msg, "\nerror: ", err)
		userMsg <- msg
	}
}
