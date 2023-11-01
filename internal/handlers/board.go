package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
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
		fmt.Println("Error decoding board", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"}, rc)
	}

	defer func() {
		cursor.Close(ctx)
	}()

	var boards []types.Board = []types.Board{}

	for cursor.Next(ctx) {
		var board types.Board
		err := cursor.Decode(&board)
		if err != nil {
			fmt.Println("Error decoding board", err)
			return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"}, rc)
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
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/boards/{short}
func (bh *BoardHandler) handleBoardShort(rc *types.RequestCtx) error {
	var pipeline bson.A
	var count int64
	var threads []bson.M
	var board *types.Board

	vars := mux.Vars(rc.Request)

	board, err := rc.Store.FindBoardByShort(vars["short"])
	if err != nil {
		return err
	}

	pipeline, err = builder.QrStrLookupThreads(board.ID, rc.Query)
	if err != nil {
		return err
	}

	count = rc.Store.CountResults("threads", append(bson.D{{Key: "board", Value: board.ID}}, rc.Query.Search...))

	threads, err = rc.Store.RunAggregation("threads", pipeline)
	if err != nil {
		return err
	}

	rc.Pagination.Update(int(count))
	rc.Records = threads
	rc.AddToResponseList("board", board)

	return ResolveResponse(rc)
}
