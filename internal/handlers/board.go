package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dd-web/opforu-server/internal/database"
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (rh *RoutingHandler) RegisterBoardRoutes(w http.ResponseWriter, r *http.Request) error {
	return nil
}

/***********************************************************************************************/
/* path: host.com/api/board
/***********************************************************************************************/
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
