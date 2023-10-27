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

/***********************************************************************************************/
/* ROOT path: host.com/api/board
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterBoardRoot(rc *types.RequestCtx) error {
	// queryCfg := utils.NewQueryConfig(r, "threads")

	switch rc.Request.Method {
	case "GET":
		return rh.handleBoardList(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/board
func (rh *RoutingHandler) handleBoardList(rc *types.RequestCtx) error {
	col := rh.Store.DB.Collection("boards")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("Error decoding board", err)
		return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"})
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
			return HandleSendJSON(rc.Writer, http.StatusInternalServerError, bson.M{"error": "unexpected server error"})
		}
		boards = append(boards, board)
	}

	return HandleSendJSON(rc.Writer, http.StatusOK, boards)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/board/{short}
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterBoardShort(rc *types.RequestCtx) error {
	// qCfg := utils.NewQueryConfig(r, "threads")

	switch rc.Request.Method {
	case "GET":
		return rh.handleBoardShort(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/board/{short}
func (rh *RoutingHandler) handleBoardShort(rc *types.RequestCtx) error {
	vars := mux.Vars(rc.Request)
	threadpipe, err := builder.QrStrLookupThreads(rh.Store.BoardIDs[vars["short"]], rc.Query)
	if err != nil {
		fmt.Println("Error building thread lookup pipeline", err)
		return err
	}

	count, err := rh.Store.CountThreadMatch(rh.Store.BoardIDs[vars["short"]], rc.Query.Search)
	if err != nil {
		fmt.Println("Error getting total record count", err)
		return err
	}

	board, err := rh.Store.FindBoardByShort(vars["short"])
	if err != nil {
		fmt.Println("Error finding board by short", err)
		return err
	}

	rc.Pagination.Update(int(count))

	threads, err := rh.Store.RunAggregation("threads", threadpipe)
	if err != nil {
		return err
	}

	rc.Records = threads

	return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"board": board, "threads": rc.Records, "paginator": rc.Pagination})
}
