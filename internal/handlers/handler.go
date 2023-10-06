package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dd-web/opforu-server/internal/database"
	"github.com/gorilla/mux"
)

type ServerHandlerFunc func(http.ResponseWriter, *http.Request) error

// All handlers are wrapped in this function to catch errors that may propagate.
// it's responsible for catching the error and sending it to the client if it's
// an error we can handle, otherwise it will send a generic error message.
// otherwise the response should be sent by the handler itself.
func FnErrWrap(f ServerHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
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

	// rh.RegisterBoardRoutes()

	// boardSubRouter := r.PathPrefix("/api/board").Subrouter()
	// r.HandleFunc("/api/board", FnErrWrap(rh.BoardHandler))

	r.Use(mux.CORSMethodMiddleware(r))

	return rh
}
