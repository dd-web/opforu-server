package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dd-web/opforu-server/internal/types"
)

type HandlerWrapperFunc func(rc *types.RequestCtx) error

// wraps a handle func with a request context and error handling
// populates request context with request details such as the account making the request (if any)
func WrapFn(f HandlerWrapperFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rc := types.NewRequestCtx(w, r)
		types.RequestLogger(rc)

		if err := f(rc); err != nil {
			fmt.Println("Error in handler:", err)
			HandleSendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}, rc)
		}
	}
}

// Handles the sending of JSON responses to the client.
// it's responsible for setting the response headers and status code as well as any data
// that needs to be sent. data should be serialized before being passed to this function.
func HandleSendJSON(w http.ResponseWriter, status int, v any, rc *types.RequestCtx) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)

	if v != nil {
		return json.NewEncoder(w).Encode(v)
	}

	return json.NewEncoder(w).Encode(map[string]string{"error": "unknown server error"})
}

// fallthrough handler for unsupported methods
func HandleUnsupportedMethod(w http.ResponseWriter, r *http.Request) error {
	return HandleSendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "unsupported method"}, nil)
}

// finalizes the response and sends it to the client
func ResolveResponse(rc *types.RequestCtx) error {
	rc.Finalize()
	return HandleSendJSON(rc.Writer, http.StatusOK, rc.ResponseData, rc)
}

// resolves an error response and sends it to the client
func ResolveResponseErr(rc *types.RequestCtx, err types.APIError) error {
	return HandleSendJSON(rc.Writer, err.Status, err.Error(), rc)
}
