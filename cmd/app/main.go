package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"github.com/dd-web/opforu-server/internal/database"
	"github.com/dd-web/opforu-server/internal/handlers"
	"github.com/dd-web/opforu-server/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	fmt.Println("Server starting up...")

	store, err := database.NewStore("opforu_local_test")
	if err != nil {
		log.Fatal(err)
	}

	err = store.HydrateBoardIDs()
	if err != nil {
		log.Fatal(err)
	}

	router := handlers.NewRoutingHandler(store)

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
func ResetDatabase(s *database.Store) {
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
