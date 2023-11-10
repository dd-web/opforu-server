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
	// this is a URL var - purely mux construct. not a header or query param
	sessionid := mux.Vars(rc.Request)["session_id"]

	// fmt.Println("looking for session", sessionid)

	session, err := rc.Store.FindSession(sessionid)
	if err != nil {
		return err
	}

	if session.IsExpired() {
		return fmt.Errorf("session is expired")
	}

	rc.AccountCtx.Session = session
	rc.AccountCtx.Account = session.Account

	return ResolveResponse(rc)
}
