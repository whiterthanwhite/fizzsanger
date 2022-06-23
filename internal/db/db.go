package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/whiterthanwhite/fizzsanger/internal/chathelper"
	"github.com/whiterthanwhite/fizzsanger/internal/helper"
)

type Conn struct {
	conn *pgx.Conn
}

func CreateConn(ctx context.Context) (*Conn, error) {
	conn, err := pgx.Connect(ctx, "postgres://localhost:5432/fizzsangerdb")
	if err != nil {
		return nil, err
	}
	return &Conn{
		conn: conn,
	}, nil
}

func (conn *Conn) Close(ctx context.Context) error {
	return conn.conn.Close(ctx)
}

func (conn *Conn) GetLastMessageID(parentCtx context.Context, chat chathelper.Chat) uint64 {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	var messageid uint64 = 0
	if err := conn.conn.QueryRow(ctx,
		`SELECT messageid FROM message WHERE chatid = $1 and userid = $2 ORDER BY messageid DESC;`,
		chat.ChatID, chat.Login).Scan(&messageid); err != nil {
		log.Println(err.Error())
		return 0
	}

	return messageid
}

func (conn *Conn) GetLastChatID(parentCtx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	var chatid string
	var err error
	if err = conn.conn.QueryRow(ctx, `select chatid from chat order by chatid desc limit 1;`).Scan(&chatid); err != nil {
		log.Println(err.Error())
		return "chat-00001", nil
	}

	if chatid, err = helper.IncStr(chatid); err != nil {
		log.Println(err.Error())
		return "", err
	}

	return chatid, nil
}

func (conn *Conn) GetUserChats(parentCtx context.Context, login string) (map[string]chathelper.Chat, error) {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	rows, _ := conn.conn.Query(ctx, `select * from chat where login = $1;`, login)
	defer rows.Close()

	userChats := make(map[string]chathelper.Chat)
	for rows.Next() {
		chat := chathelper.Chat{}
		if err := rows.Scan(&chat.ChatID, &chat.ChatType, &chat.CreationDateTime, &chat.Login); err != nil {
			return nil, err
		}
		userChats[chat.ChatID] = chat
	}

	return userChats, nil
}

func (conn *Conn) SaveChat(parentCtx context.Context, chat chathelper.Chat) {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	if err := conn.conn.QueryRow(ctx, `insert into chat values($1, $2, $3, $4)`,
		chat.ChatID, chat.ChatType, chat.CreationDateTime.Format("2006-01-02 15:04:05-0700"), chat.Login).
		Scan(); err != nil {
		log.Println(err.Error())
		return
	}
}

func (conn *Conn) SaveUserMessage(parentCtx context.Context, chat chathelper.Chat, msg []byte) {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	messageid := conn.GetLastMessageID(ctx, chat)
	messageid += 1
	if err := conn.conn.QueryRow(ctx, `insert into message values($1, $2, $3, $4, $5, $6, $7)`,
		chat.ChatID, chat.Login, messageid, string(msg),
		time.Now().Format("2006-01-02 15:04:05-0700"), 0, 0).
		Scan(); err != nil {
		log.Println(err.Error())
		return
	}
}
