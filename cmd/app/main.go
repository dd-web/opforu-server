package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/dd-web/opforu-server/internal/database"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/dd-web/opforu-server/internal/utils"
)

type (
	Account  types.Account
	Article  types.Article
	Board    types.Board
	Identity types.Identity
	Post     types.Post
	Thread   types.Thread
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

	queryConfig, err := utils.NewQueryConfig(nil)
	if err != nil {
		log.Fatal(err)
	}

	storeAgg, err := store.RunAggregation("boards", queryConfig)
	if err != nil {
		log.Fatal(err)
	}

	var boardRes types.Board = types.Board{}
	err = types.UnmarshalBoard(storeAgg[0], &boardRes)
	if err != nil {
		log.Fatal(err)
	}

	// doc, err := bson.Marshal(storeAgg[0])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = bson.Unmarshal(doc, &boardRes)

	// if err = types.UnMarshalBoard(storeAgg[0], &boardRes); err != nil {
	// 	log.Fatal(err)
	// }

	fmt.Println("BOARD!!!:", boardRes)
	fmt.Println("board ID", boardRes.ID)
	fmt.Println("board Title", boardRes.Title)

	// to get rid of compile errors for now
	fmt.Println("Store: ", store)

	router := mux.NewRouter()

	// to get rid of compile errors for now
	fmt.Println("Router: ", router)

	// router.HandleFunc("/api/v1/account", HandleWrapperFunc(handlers.AccountRoot(store)))

	fmt.Println("Server shutting down...")
}

// func HandleWrapperFunc(f handlers.ServerHandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if err := f(w, r); err != nil {
// 			RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
// 		}
// 	}
// }

// func RespondJSON(w http.ResponseWriter, status int, v any) error {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")

// 	if v != nil {
// 		return json.NewEncoder(w).Encode(v)
// 	}
// 	return json.NewEncoder(w).Encode(map[string]string{"error": "unknown server error"})
// }

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
