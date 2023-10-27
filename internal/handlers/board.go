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
	rc *types.RequestCtx
	rh *types.RoutingHandler
}

func InitBoardHandler(rh *types.RoutingHandler) *BoardHandler {
	return &BoardHandler{
		rh: rh,
	}
}

func (bh *BoardHandler) UpdateCtx(rc *types.RequestCtx) {
	bh.rc = rc
}

/***********************************************************************************************/
/* ROOT path: host.com/api/board
/***********************************************************************************************/
func (bh *BoardHandler) RegisterBoardRoot(rc *types.RequestCtx) error {
	bh.UpdateCtx(rc)
	// queryCfg := utils.NewQueryConfig(r, "threads")

	switch rc.Request.Method {
	case "GET":
		return bh.handleBoardList(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/board
func (bh *BoardHandler) handleBoardList(rc *types.RequestCtx) error {
	bh.UpdateCtx(rc)
	col := bh.rh.Store.DB.Collection("boards")
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
func (bh *BoardHandler) RegisterBoardShort(rc *types.RequestCtx) error {
	bh.UpdateCtx(rc)
	// qCfg := utils.NewQueryConfig(r, "threads")

	switch rc.Request.Method {
	case "GET":
		return bh.handleBoardShort(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/board/{short}
func (bh *BoardHandler) handleBoardShort(rc *types.RequestCtx) error {
	bh.UpdateCtx(rc)
	vars := mux.Vars(rc.Request)
	threadpipe, err := builder.QrStrLookupThreads(bh.rh.Store.BoardIDs[vars["short"]], rc.Query)
	if err != nil {
		fmt.Println("Error building thread lookup pipeline", err)
		return err
	}

	count, err := bh.rh.Store.CountThreadMatch(bh.rh.Store.BoardIDs[vars["short"]], rc.Query.Search)
	if err != nil {
		fmt.Println("Error getting total record count", err)
		return err
	}

	board, err := bh.rh.Store.FindBoardByShort(vars["short"])
	if err != nil {
		fmt.Println("Error finding board by short", err)
		return err
	}

	rc.Pagination.Update(int(count))

	threads, err := bh.rh.Store.RunAggregation("threads", threadpipe)
	if err != nil {
		return err
	}

	rc.Records = threads

	return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"board": board, "threads": rc.Records, "paginator": rc.Pagination})
}
