package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/fizzsanger/internal/config"
	"github.com/whiterthanwhite/fizzsanger/internal/handlers"
)

func StartServer(parentCtx context.Context, authServer *http.Server) {
	_, cancel := context.WithCancel(parentCtx)
	if err := authServer.ListenAndServe(); err != nil {
		log.Println(err.Error())
		cancel()
	}
	cancel()
}

func main() {
	log.Printf("Server start at %s\n", time.Now().String())
	conf := config.GetConf()
	if conf == nil {
		log.Fatal("Server configuration is not setted!")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chiRouter := chi.NewRouter()
	chiRouter.Route("/", func(cr chi.Router) {
		cr.Get("/", handlers.GetRegisterPage)
		cr.Post("/register", handlers.UserRegister(conf))
		cr.Post("/login", handlers.UserLogin(conf))
	})

	authServer := &http.Server{
		Addr:    ":8080",
		Handler: chiRouter,
	}

	go StartServer(ctx, authServer)

	log.Println("Server started successfully")
	<-ctx.Done()
	if err := authServer.Close(); err != nil {
		log.Printf("Server closed with error: %s\n", err.Error())
	}
	log.Println("Server closed successfully")
}
