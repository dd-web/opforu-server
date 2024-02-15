package handlers

import (
	"encoding/json"
	"io"
	"time"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/gorilla/mux"
)

type ThreadHandler struct {
	rh *types.RoutingHandler
}

func InitThreadHandler(rh *types.RoutingHandler) *ThreadHandler {
	return &ThreadHandler{
		rh: rh,
	}
}

/***********************************************************************************************/
/* ROOT path: host.com/api/threads/{slug}
/***********************************************************************************************/
func (th *ThreadHandler) RegisterThreadRoot(rc *types.RequestCtx) error {
	rc.UpdateStore(th.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return th.handleThreadRoot(rc)
	case "POST":
		return th.handleThreadReply(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/threads/{slug}
func (th *ThreadHandler) handleThreadRoot(rc *types.RequestCtx) error {
	vars := mux.Vars(rc.Request)

	pipeline := builder.QrStrEntireThread(vars["slug"], rc.Query)

	result, err := th.rh.Store.RunAggregation("threads", pipeline)
	if err != nil {
		return err
	}

	rc.AddToResponseList("thread", result[0])
	return ResolveResponse(rc)
}

// METHOD: POST
// PATH: host.com/api/threads/{slug}
func (th *ThreadHandler) handleThreadReply(rc *types.RequestCtx) error {
	if rc.UnresolvedAccount {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	// invoke necessary data and dependencies
	vars := mux.Vars(rc.Request)

	ts := time.Now().UTC()
	details := types.NewRUMPost()

	thread, err := th.rh.Store.FindThreadBySlug(vars["slug"])
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	board, err := th.rh.Store.FindBoardByObjectID(thread.Board)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = json.Unmarshal(body, &details)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	identity, err := th.rh.Store.ResolveIdentity(rc.AccountCtx.Account.ID, thread.ID)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	// dependency injection
	newPostAssets := []*types.Asset{}
	newPostAssetInterfaces := []interface{}{}

	// no idea why this check is necessary but it is
	// there is otherwise no stack trace information and we regardless receive 200 response? weird.
	if len(details.Assets) > 0 {
		for _, v := range details.Assets {
			a := types.NewAsset(v.SourceID, rc.AccountCtx.Account.ID)
			a.FileName = v.FileName
			a.Description = v.Description
			a.Tags = v.Tags
			newPostAssets = append(newPostAssets, a)
		}
	}

	tstore := types.NewTemplateStore()
	str, err := tstore.Parse(details.Content)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	post := types.NewPost()

	post.PostNumber = board.PostRef + 1
	post.Creator = identity.ID
	post.Body = str
	post.Board = board.ID
	post.Thread = thread.ID
	thread.Posts = append(thread.Posts, post.ID)
	board.PostRef++

	thread.UpdatedAt = &ts
	board.UpdatedAt = &ts

	if len(details.Assets) > 0 {
		for _, v := range newPostAssets {
			post.Assets = append(post.Assets, v.ID)
			newPostAssetInterfaces = append(newPostAssetInterfaces, v)
		}
		err = rc.Store.SaveNewMulti(newPostAssetInterfaces, "assets")
		if err != nil {
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}
	}

	err = rc.Store.SaveNewSingle(post, "posts")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = rc.Store.UpdateThread(thread)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = rc.Store.UpdateBoard(board)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	rc.AddToResponseList("post_number", post.PostNumber)
	return ResolveResponse(rc)
}
