package hub

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
	"github.com/whiterthanwhite/fizzsanger/internal/chathelper"
	"github.com/whiterthanwhite/fizzsanger/internal/db"
)

type Client struct {
	Login   string
	Token   *jwt.Token
	Conn    *websocket.Conn
	Message chan []byte
}

func (client *Client) GetMessages() {
	log.Println(client)
	defer client.Conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		msgType, msg, err := client.Conn.ReadMessage()
		if err != nil {
			log.Println(err.Error())
			return
		}
		log.Println("\nMessage type: ", msgType, "\nmessage: ", string(msg), "\nerror: ", err)
		decodedMsg, err := base64.StdEncoding.DecodeString(string(msg))
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Println(string(decodedMsg))
		client.Message <- decodedMsg
		received_message := ParseMessage(<-client.Message)
		if received_message == nil {
			continue
		}
		if len(received_message.Message) == 0 {
			log.Println("empty message")
			continue
		}

		dbconn, err := db.CreateConn(ctx)
		if err != nil {
			log.Println(err.Error())
			break
		}

		var chat chathelper.Chat
		var ok bool

		if received_message.ChatID == "" {
			received_message.ChatID, _ = dbconn.GetLastChatID(ctx)
		}
		chats, _ := dbconn.GetUserChats(ctx, client.Login)
		if chat, ok = chats[received_message.ChatID]; !ok {
			log.Println(received_message.ChatID, " exists: ", ok)
			chat = chathelper.CreateChat(received_message.ChatID, 0, client.Login)
			dbconn.SaveChat(ctx, chat)
		}

		dbconn.SaveUserMessage(ctx, chat, received_message.Message)

		if err = dbconn.Close(ctx); err != nil {
			log.Println(err.Error())
		}
	}
}

func (client *Client) SendMessages(msg []byte) {
	if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Println(err.Error())
		return
	}
}

func ParseMessage(message []byte) *chathelper.Message {
	msg := &chathelper.Message{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println(err.Error())
		return nil
	}
	return msg
}

func (client *Client) GetChats() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := db.CreateConn(ctx)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer conn.Close(ctx)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		userChats, err := conn.GetUserChats(context.Background(), client.Login)
		if err != nil {
			log.Println(err.Error())
		}
		if userChats != nil {
			chats := chathelper.Chats{
				Chats: userChats,
			}
			mc, err := chats.MarshallChats()
			if err == nil {
				log.Println(string(mc))
				dstMsg := make([]byte, base64.StdEncoding.EncodedLen(len(mc)))
				base64.StdEncoding.Encode(dstMsg, mc)
				client.SendMessages(dstMsg)
			}
		}
	}
}

func (client *Client) EndConnection() {
	if err := client.Conn.Close(); err != nil {
		log.Println(err.Error())
	}
	log.Println(client.Login, ": connection closed")
}

type Hub struct {
	Clients map[string]*Client
}

func CreateHub() *Hub {
	return &Hub{
		Clients: make(map[string]*Client, 0),
	}
}
