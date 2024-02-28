// This handler is an intermediate handler for use by the front end server to communicate with
// it's for account/session lookup and other services.
// The client should never be able to access this handler directly.
// since these handlers are very specific they won't have switch registers for their methods
// this also means we should immediately update the store on every handler call

package handlers

import (
	"fmt"
	"strconv"

	"github.com/dd-web/opforu-server/internal/builder"
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
		fmt.Println("session expired")
		_ = rc.Store.DeleteSingle(session.ID, "sessions")
		return fmt.Errorf("session expired")
	}

	rc.AccountCtx.Session = session
	rc.AccountCtx.Account = session.Account

	return ResolveResponse(rc)
}

// METHOD: GET
// PATH: /api/internal/post/{thread_slug}/{post_number}
// retreives a post by first looking up the thread, then finding the post with that thread_id and post_number
func (ih *InternalHandler) HandleGetPost(rc *types.RequestCtx) error {
	rc.UpdateStore(ih.rh.Store)
	vars := mux.Vars(rc.Request)

	thread, err := rc.Store.FindThreadBySlug(vars["thread_slug"])
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("thread"))
	}

	board, err := rc.Store.FindBoardByObjectID(thread.Board)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("post's board"))
	}

	postNum, err := strconv.Atoi(vars["post_number"])
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("post"))
	}

	if postNum < 1 {
		return ResolveResponseErr(rc, types.ErrorNotFound("post"))
	}

	p, err := rc.Store.RunAggregation("posts", builder.QrStrLookupPost(thread.ID, postNum, thread.Slug, board.Short))
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("post"))
	}

	if len(p) == 0 {
		return ResolveResponseErr(rc, types.ErrorNotFound("post"))
	}

	if p[0] == nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	rc.AddToResponseList("post", p[0])

	return ResolveResponse(rc)
}

// METHOD: GET
// PATH: /api/internal/thread/{board_short}/{thread_slug}
func (ih *InternalHandler) HandleGetThread(rc *types.RequestCtx) error {
	rc.UpdateStore(ih.rh.Store)
	vars := mux.Vars(rc.Request)

	thread, err := rc.Store.FindThreadBySlug(vars["thread_slug"])
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("thread"))
	}

	board, err := rc.Store.FindBoardByObjectID(thread.Board)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("thread's board"))
	}

	t, err := rc.Store.RunAggregation("threads", builder.QrStrLookupThread(thread.Slug, board.ID, board.Short))
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("thread"))
	}

	if len(t) == 0 || t[0] == nil {
		return ResolveResponseErr(rc, types.ErrorNotFound("thread"))
	}

	rc.AddToResponseList("thread", t[0])

	return ResolveResponse(rc)
}
