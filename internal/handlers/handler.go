package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dd-web/opforu-server/internal/database"
	"github.com/gorilla/mux"
)

type HandlerWrapperFunc func(http.ResponseWriter, *http.Request) error

// All handlers are wrapped in this function to catch errors that may propagate.
// it's responsible for catching the error and sending it to the client if it's
// an error we can handle, otherwise it will send a generic error message.
// otherwise the response should be sent by the handler itself.
func WrapFn(f HandlerWrapperFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			fmt.Println("Error in handler:", err)
			HandleSendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
}

// Handles the sending of JSON responses to the client.
// it's responsible for setting the response headers and status code as well as any data
// that needs to be sent. data should be serialized before being passed to this function.
func HandleSendJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)

	if v != nil {
		return json.NewEncoder(w).Encode(v)
	}
	return json.NewEncoder(w).Encode(map[string]string{"error": "unknown server error"})
}

type RoutingHandler struct {
	Router *mux.Router
	Store  *database.Store
}

// Creates a new route handler constructor with access to the store for database operations
// the returned handler's router should be used as the main router for the server.
func NewRoutingHandler(s *database.Store) *RoutingHandler {
	r := mux.NewRouter()
	rh := &RoutingHandler{
		Router: r,
		Store:  s,
	}

	fmt.Println("Registering routes...")

	rh.Router.HandleFunc("/api/boards/{short}", WrapFn(rh.RegisterBoardShort))
	rh.Router.HandleFunc("/api/boards", WrapFn(rh.RegisterBoardRoot))

	rh.Router.Use(mux.CORSMethodMiddleware(rh.Router))

	return rh
}

func HandleUnsupportedMethod(w http.ResponseWriter, r *http.Request) error {
	return HandleSendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "unsupported method"})
}
