package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/database"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/dd-web/opforu-server/internal/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

/***********************************************************************************************/
/* path: host.com/api/board
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterBoardRoot(w http.ResponseWriter, r *http.Request) error {
	queryCfg := utils.NewQueryConfig(r, "threads")

	switch r.Method {
	case "GET":
		return rh.handleBoardList(w, r, queryCfg)
	default:
		return HandleUnsupportedMethod(w, r)
	}
}

func getBoardList(s *database.Store) ([]types.Board, error) {
	col := s.DB.Collection("boards")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
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
		}
		boards = append(boards, board)
	}

	if err = cursor.Err(); err != nil {
		log.Println("Error decoding board", err)
		return nil, err
	}

	return boards, nil
}

// GET: host.com/api/board
func (rh *RoutingHandler) handleBoardList(w http.ResponseWriter, r *http.Request, q *utils.QueryConfig) error {
	boards, err := getBoardList(rh.Store)
	if err != nil {
		return err
	}

	return HandleSendJSON(w, http.StatusOK, boards)
}

/***********************************************************************************************/
/* path: host.com/api/board/{short}
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterBoardShort(w http.ResponseWriter, r *http.Request) error {
	qCfg := utils.NewQueryConfig(r, "threads")

	switch r.Method {
	case "GET":
		return rh.handleBoardShort(w, r, qCfg)
	default:
		return HandleUnsupportedMethod(w, r)
	}
}

// GET: host.com/api/board/{short}
func (rh *RoutingHandler) handleBoardShort(w http.ResponseWriter, r *http.Request, q *utils.QueryConfig) error {
	vars := mux.Vars(r)
	threadpipe, err := builder.QrStrLookupThreads(rh.Store.BoardIDs[vars["short"]], q)
	if err != nil {
		fmt.Println("Error building thread lookup pipeline", err)
		return err
	}

	count, err := rh.Store.CountThreadMatch(rh.Store.BoardIDs[vars["short"]], q.Search)
	if err != nil {
		fmt.Println("Error getting total record count", err)
		return err
	}

	board, err := rh.Store.FindBoardByShort(vars["short"])
	if err != nil {
		fmt.Println("Error finding board by short", err)
		return err
	}

	q.PageInfo.Update(int(count))

	threads, err := rh.Store.RunAggregation("threads", threadpipe)
	if err != nil {
		return err
	}

	q.PageInfo.Records = threads

	return HandleSendJSON(w, http.StatusOK, bson.M{"board": board, "data": q.PageInfo})
}
