// This handler is an intermediate handler for use by the front end server to communicate with
// it's for account/session lookup and other services.
// The client should never be able to access this handler directly.
// since these handlers are very specific they won't have switch registers for their methods
// this also means we should immediately update the store on every handler call

package handlers

import (
	"fmt"

	"github.com/dd-web/opforu-server/internal/types"
	"github.com/gorilla/mux"
)

type InternalHandler struct {
	rh *types.RoutingHandler
}

func InitInternalHandlers(rh *types.RoutingHandler) *InternalHandler {
	return &InternalHandler{
		rh: rh,
	}
}

// METHOD: GET
// PATH: /api/internal/session/{session_id}
// retreives a session from it's id
func (ih *InternalHandler) HandleGetSession(rc *types.RequestCtx) error {
	rc.UpdateStore(ih.rh.Store)
	vars := mux.Vars(rc.Request)

	fmt.Println("looking for session", vars["session_id"])

	return nil
}
