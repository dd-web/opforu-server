package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/dd-web/opforu-server/internal/handlers"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/dd-web/opforu-server/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	fmt.Println("Server starting up...")

	store, err := types.NewStore("opforu_local_test")
	if err != nil {
		log.Fatal(err)
	}

	err = store.HydrateBoardIDs()
	if err != nil {
		log.Fatal(err)
	}

	router := types.NewRoutingHandler(store)

	handler_account := handlers.InitAccountHandlers(router)
	handler_board := handlers.InitBoardHandler(router)
	handler_thread := handlers.InitThreadHandler(router)

	fmt.Println("Registering handlers...")

	// account
	router.Router.HandleFunc("/api/account/login", handlers.WrapFn(handler_account.RegisterAccountLogin))
	router.Router.HandleFunc("/api/account/register", handlers.WrapFn(handler_account.RegisterAccountRegister))
	router.Router.HandleFunc("/api/account/me", handlers.WrapFn(handler_account.RegisterAccountMe))
	router.Router.HandleFunc("/api/account", handlers.WrapFn(handler_account.RegisterAccountRoot))

	// boards
	router.Router.HandleFunc("/api/boards/{short}", handlers.WrapFn(handler_board.RegisterBoardShort))
	router.Router.HandleFunc("/api/boards", handlers.WrapFn(handler_board.RegisterBoardRoot))

	// threads
	router.Router.HandleFunc("/api/threads/{slug}", handlers.WrapFn(handler_thread.RegisterThreadRoot))

	// request config
	router.Router.Use(mux.CORSMethodMiddleware(router.Router))

	srv := &http.Server{
		Handler:      router.Router,
		Addr:         "127.0.0.1:3001",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
	fmt.Println("Server shutting down...")
}

// ResetDatabase drops the database and recreates it
func ResetDatabase(s *types.Store) {
	if utils.IsProdEnv() {
		log.Fatal("Will not reset database in production environment.")
	}

	// drop db
	err := s.DB.Drop(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	cols := []string{"accounts", "articles", "boards", "identities", "posts", "threads"}

	// create collections
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, name := range cols {
		err := s.DB.CreateCollection(ctx, name)
		if err != nil {
			fmt.Println("Error creating collection: ", name)
			continue
		}
	}
}
