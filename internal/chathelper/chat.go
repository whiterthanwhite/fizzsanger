package chathelper

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
)

type (
	Chats struct {
		Chats map[string]Chat `json:"chats"`
	}

	Chat struct {
		ChatID           string    `json:"chatid"`
		ChatType         uint8     `json:"chat_type"`
		CreationDateTime time.Time `json:"creation_datetime"`
		Login            string    `json:"login"`
	}

	Message struct {
		ChatID       string `json:"chatid"`
		Login        string `json:"login"`
		Message      []byte `json:"message"`
		ReplyMessage int    `json:"reply_message"`
	}
)

func (m *Message) UnmarshalJSON(data []byte) error {
	type temp struct {
		ChatID       string `json:"chatid"`
		Message      string `json:"message"`
		ReplyMessage string `json:"reply_message"`
	}
	t := temp{}
	if err := json.Unmarshal(data, &t); err != nil {
		log.Println(err.Error())
		return err
	}
	replyMessageID, err := strconv.ParseInt(t.ReplyMessage, 0, 64)
	if err != nil {
		return nil
	}

	m.ChatID = t.ChatID
	m.Message = []byte(t.Message)
	m.ReplyMessage = int(replyMessageID)

	return nil
}

func CreateChat(chatid string, chatType uint8, login string) Chat {
	chat := Chat{
		ChatID:           chatid,
		ChatType:         chatType,
		CreationDateTime: time.Now(),
		Login:            login,
	}
	return chat
}

func (chats Chats) MarshallChats() ([]byte, error) {
	jsonChats, err := json.Marshal(&chats)
	if err != nil {
		return nil, err
	}
	return jsonChats, nil
}
