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

	// disable cache until fully implemented
	// err = store.HydrateCache()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	handler := types.NewRoutingHandler(store)

	handler_account := handlers.InitAccountHandlers(handler)
	handler_article := handlers.InitArticleHandler(handler)
	handler_asset := handlers.InitAssetHandler(handler)
	handler_board := handlers.InitBoardHandler(handler)
	handler_thread := handlers.InitThreadHandler(handler)
	handler_internal := handlers.InitInternalHandlers(handler)

	fmt.Println("Registering handlers...")

	// account
	handler.Router.HandleFunc("/api/account/posts", handlers.WrapFn(handler_account.RegisterAccountPosts))
	handler.Router.HandleFunc("/api/account/logout", handlers.WrapFn(handler_account.RegisterAccountLogout))
	handler.Router.HandleFunc("/api/account/login", handlers.WrapFn(handler_account.RegisterAccountLogin))
	handler.Router.HandleFunc("/api/account/register", handlers.WrapFn(handler_account.RegisterAccountRegister))
	handler.Router.HandleFunc("/api/account", handlers.WrapFn(handler_account.RegisterAccountRoot))

	// articles
	handler.Router.HandleFunc("/api/articles", handlers.WrapFn(handler_article.RegisterArticleRoot))
	handler.Router.HandleFunc("/api/articles/{slug}", handlers.WrapFn(handler_article.RegisterArticleSlug))

	// assets
	handler.Router.HandleFunc("/api/assets", handlers.WrapFn(handler_asset.RegisterAssetRoot))

	// boards
	handler.Router.HandleFunc("/api/boards/{short}", handlers.WrapFn(handler_board.RegisterBoardShort))
	handler.Router.HandleFunc("/api/boards", handlers.WrapFn(handler_board.RegisterBoardRoot))

	// threads
	handler.Router.HandleFunc("/api/threads/{slug}", handlers.WrapFn(handler_thread.RegisterThreadRoot))

	// internal server routes
	handler.Router.HandleFunc("/api/internal/session/{session_id}", handlers.WrapFn(handler_internal.HandleGetSession))
	handler.Router.HandleFunc("/api/internal/post/{thread_slug}/{post_number}", handlers.WrapFn(handler_internal.HandleGetPost))
	handler.Router.HandleFunc("/api/internal/thread/{board_short}/{thread_slug}", handlers.WrapFn(handler_internal.HandleGetThread))

	// request config
	handler.Router.Use(mux.CORSMethodMiddleware(handler.Router))

	srv := &http.Server{
		Handler:      handler.Router,
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
