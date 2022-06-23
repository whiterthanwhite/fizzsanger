package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/whiterthanwhite/fizzsanger/internal/chathelper"
	"github.com/whiterthanwhite/fizzsanger/internal/config"
	"github.com/whiterthanwhite/fizzsanger/internal/helper"
)

type Conn struct {
	conn *pgx.Conn
}

func CreateConn(ctx context.Context, conf *config.Conf) (*Conn, error) {
	conn, err := pgx.Connect(ctx, conf.DBAddress)
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
		return "chat-00000", nil
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

func (conn *Conn) GetLastUserID(parentCtx context.Context) string {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	var userid string
	if err := conn.conn.QueryRow(ctx, `select userid from user_tab order by userid desc limit 1;`).Scan(&userid); err != nil {
		log.Println(err.Error())
		return `user-00001`
	}

	return userid
}

func (conn *Conn) IsLoginExist(parentCtx context.Context, login string) bool {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	ct, err := conn.conn.Exec(ctx, `select * from user_tab where login = $1 limit 1;`, login)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	if ct.RowsAffected() == 0 {
		return false
	}

	return true
}

func (conn *Conn) SaveUser(parentCtx context.Context, userid, login string, pass []byte) bool {
	ctx, cancel := context.WithTimeout(parentCtx, time.Second)
	defer cancel()

	if err := conn.conn.QueryRow(ctx,
		`insert into user_tab (userid, login, password, creation_datetime) values ($1, $2, $3, $4)`,
		userid, login, pass, time.Now().Format("2006-01-02 15:04:05-0700")).
		Scan(); err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}
