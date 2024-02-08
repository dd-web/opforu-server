package handlers

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BoardHandler struct {
	rh *types.RoutingHandler
}

func InitBoardHandler(rh *types.RoutingHandler) *BoardHandler {
	return &BoardHandler{
		rh: rh,
	}
}

/***********************************************************************************************/
/* ROOT path: host.com/api/boards
/***********************************************************************************************/
func (bh *BoardHandler) RegisterBoardRoot(rc *types.RequestCtx) error {
	rc.UpdateStore(bh.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return bh.handleBoardList(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/boards
func (bh *BoardHandler) handleBoardList(rc *types.RequestCtx) error {
	col := rc.Store.DB.Collection("boards")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	defer func() {
		cursor.Close(ctx)
	}()

	var boards []types.Board = []types.Board{}

	for cursor.Next(ctx) {
		var board types.Board
		err := cursor.Decode(&board)
		if err != nil {
			return ResolveResponseErr(rc, types.ErrorUnexpected())
		}
		boards = append(boards, board)
	}

	rc.AddToResponseList("boards", boards)
	return ResolveResponse(rc)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/boards/{short}
/***********************************************************************************************/
func (bh *BoardHandler) RegisterBoardShort(rc *types.RequestCtx) error {
	rc.UpdateStore(bh.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return bh.handleBoardShort(rc)
	case "POST":
		return bh.handleNewThread(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/boards/{short}
func (bh *BoardHandler) handleBoardShort(rc *types.RequestCtx) error {
	vars := mux.Vars(rc.Request)
	var board *types.Board

	board, err := rc.Store.FindBoardByShort(vars["short"])
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	pipeline, err := builder.QrStrLookupThreads(board.ID, rc.Query)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	count := rc.Store.CountResults("threads", append(bson.D{{Key: "board", Value: board.ID}}, rc.Query.Search...))

	threads, err := rc.Store.RunAggregation("threads", pipeline)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	rc.Pagination.Update(int(count))
	rc.Records = threads
	rc.AddToResponseList("board", board)

	return ResolveResponse(rc)
}

// METHOD: POST
// PATH: host.com/api/boards/{short}
func (bh *BoardHandler) handleNewThread(rc *types.RequestCtx) error {
	if rc.UnresolvedAccount {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	// invoke necessary data and dependencies
	vars := mux.Vars(rc.Request)
	board, err := rc.Store.FindBoardByShort(vars["short"])
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	details := types.NewRUMThread()

	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = json.Unmarshal(body, &details)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	// dependency injection
	newThreadAssets := []*types.Asset{}
	newThreadAssetInterfaces := []interface{}{}

	for _, v := range details.Assets {
		a := types.NewAsset(v.SourceID, rc.AccountCtx.Account.ID)
		a.FileName = v.FileName
		a.Description = v.Description
		a.Tags = v.Tags
		newThreadAssets = append(newThreadAssets, a)
	}

	newIdentity := types.NewIdentity()
	newIdentity.Account = rc.AccountCtx.Account.ID
	newIdentity.Role = types.ThreadRoleCreator

	thread := types.NewThread()

	newIdentity.Thread = thread.ID
	thread.Title = details.Title
	thread.Body = details.Content
	thread.Board = board.ID
	thread.Creator = newIdentity.ID
	thread.Mods = []primitive.ObjectID{newIdentity.ID}

	for _, v := range newThreadAssets {
		thread.Assets = append(thread.Assets, v.ID)
		newThreadAssetInterfaces = append(newThreadAssetInterfaces, v)
	}

	// save & update associated docs
	err = rc.Store.SaveNewMulti(newThreadAssetInterfaces, "assets")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = rc.Store.SaveNewSingle(newIdentity, "identities")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = rc.Store.SaveNewSingle(thread, "threads")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	rc.AddToResponseList("thread_id", thread.Slug)
	return ResolveResponse(rc)
}
