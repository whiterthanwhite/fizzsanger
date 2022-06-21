package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/fizzsanger/internal/config"
	"github.com/whiterthanwhite/fizzsanger/internal/handlers"
	"github.com/whiterthanwhite/fizzsanger/internal/hub"
)

func ShowHubUsers(parentCtx context.Context, h *hub.Hub) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for client := range h.Clients {
				log.Println(client)
			}
		case <-ctx.Done():
			return
		}
	}
}

func startServer(parentCtx context.Context, authServer *http.Server) {
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

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	hub := hub.CreateHub()

	chiRouter := chi.NewRouter()
	chiRouter.Route("/", func(cr chi.Router) {
		cr.Get("/", handlers.GetMessengerPage)
		cr.Get("/chat", handlers.ConnectToChat(hub, conf))
	})

	authServer := &http.Server{
		Addr:    ":8081",
		Handler: chiRouter,
	}

	go startServer(ctx, authServer)
	go ShowHubUsers(ctx, hub)

	log.Println("Server started successfully")
	<-ctx.Done()
	if err := authServer.Close(); err != nil {
		log.Printf("Server closed with error: %s\n", err.Error())
	}
	log.Println("Server closed successfully")
}
