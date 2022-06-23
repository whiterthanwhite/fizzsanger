package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4"

	"github.com/whiterthanwhite/fizzsanger/internal/auth"
	"github.com/whiterthanwhite/fizzsanger/internal/config"
	"github.com/whiterthanwhite/fizzsanger/internal/db"
	"github.com/whiterthanwhite/fizzsanger/internal/helper"
	"github.com/whiterthanwhite/fizzsanger/internal/hub"
)

const (
	wrongMethodErr   = `Wrong request method!`
	occupiedLoginErr = `Login is occupied!`
)

// authserver func handlers
func GetRegisterPage(rw http.ResponseWriter, r *http.Request) {
	log.Println("Handler: GetRegisterPage")
	if r.Method != http.MethodGet {
		http.Error(rw, wrongMethodErr, http.StatusMethodNotAllowed)
		return
	}

	currPath, err := os.Getwd()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = os.Chdir(`../../internal/template`); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	templateByte, err := os.ReadFile(currPath + `/index.html`)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	t := template.New(`registerpage`)
	if t, err = t.Parse(string(templateByte)); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = t.Execute(rw, nil); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UserRegister(conf *config.Conf) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Println("UserRegister")

		log.Println("Request Content-Type: : ", r.Header.Get("Content-Type"))
		headContentType := r.Header.Get("Content-Type")
		if headContentType != "application/json" {
			http.Error(rw, "Wrong Content-Type", http.StatusBadRequest)
			return
		}

		var err error
		var rBody []byte

		if rBody, err = io.ReadAll(r.Body); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println(string(rBody))
		if err = r.Body.Close(); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(rBody) == 0 {
			http.Error(rw, "empty body", http.StatusBadRequest)
			return
		}

		userCredentials := auth.UserCredentials{}
		if err = json.Unmarshal(rBody, &userCredentials); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		// Hash password >>
		h := sha256.New()
		if _, err := h.Write([]byte(userCredentials.Password)); err != nil {
			log.Println(err.Error())
			http.Error(rw, "Register error", http.StatusInternalServerError)
			return
		}
		passHash := h.Sum(nil)
		// Hash password <<

		// database >>
		conn, err := db.CreateConn(r.Context(), conf)
		if err != nil {
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
		defer conn.Close(r.Context())

		if conn.IsLoginExist(r.Context(), userCredentials.Login) {
			http.Error(rw, occupiedLoginErr, http.StatusConflict)
			return
		}

		userid := conn.GetLastUserID(r.Context())
		userid, err = helper.IncStr(userid)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if !conn.SaveUser(r.Context(), userid, userCredentials.Login, passHash) {
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
		// database <<

		rw.WriteHeader(http.StatusOK)
	}
}

func UserLogin(conf *config.Conf) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "Wrong Context-Type", http.StatusBadRequest)
			return
		}

		var err error
		var rBody []byte

		if rBody, err = io.ReadAll(r.Body); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = r.Body.Close(); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(rBody) == 0 {
			http.Error(rw, "empty body", http.StatusBadRequest)
			return
		}

		userCredentials := auth.UserCredentials{}
		if err = json.Unmarshal(rBody, &userCredentials); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		conn, err := pgx.Connect(r.Context(), "postgres://localhost:5432/fizzsangerdb")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close(r.Context())

		var password []byte
		if err := conn.QueryRow(r.Context(), `SELECT password FROM user_tab WHERE login = $1;`,
			userCredentials.Login).Scan(&password); err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		h := sha256.New()
		if _, err := h.Write([]byte(userCredentials.Password)); err != nil {
			log.Println(err.Error())
			http.Error(rw, "Login error", http.StatusInternalServerError)
			return
		}
		passHash := h.Sum(nil)

		if string(passHash) != string(password) {
			http.Error(rw, "Authentification error", http.StatusUnauthorized)
			return
		}

		// create token
		claims := auth.CreateCustomClaims(userCredentials.Login, conf)
		token, err := auth.CreateToken(claims, conf)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		// set cookie
		cookieToken := http.Cookie{
			Name:     "auth-token",
			Value:    token,
			HttpOnly: true,
		}
		http.SetCookie(rw, &cookieToken)

		rw.WriteHeader(http.StatusOK)
	}
}

// mainserver func handlers
func GetMessengerPage(rw http.ResponseWriter, r *http.Request) {
	log.Println("Handler: GetRegisterPage")
	if r.Method != http.MethodGet {
		http.Error(rw, wrongMethodErr, http.StatusMethodNotAllowed)
		return
	}

	currPath, err := os.Getwd()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = os.Chdir(`../../internal/template`); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	templateByte, err := os.ReadFile(currPath + `/messagepage.html`)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	t := template.New(`mainpage`)
	if t, err = t.Parse(string(templateByte)); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = t.Execute(rw, nil); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ConnectToChat(h *hub.Hub, conf *config.Conf) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Println("ConnectToChat")

		cookieToken, err := r.Cookie("auth-token")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		claims := auth.MyCustomClaims{}
		token, err := auth.ParseToken(cookieToken.Value, &claims, conf)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if !token.Valid {
			http.Error(rw, "Token is not valid", http.StatusUnauthorized)
			return
		}

		wsupgrader := websocket.Upgrader{}
		wsconn, err := wsupgrader.Upgrade(rw, r, nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		client := &hub.Client{
			Login:   claims.Login,
			Token:   token,
			Conn:    wsconn,
			Message: make(chan []byte, 1),
		}
		h.Clients[client.Login] = client

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		go client.GetChats(conf)
		go client.GetMessages(conf)

		<-ctx.Done()
		log.Println("End ConnectToChat")
	}
}
