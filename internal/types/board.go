package types

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Board struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	Title       string `bson:"title" json:"title"`
	Short       string `bson:"short" json:"short"` // short name for the board (used in URLs)
	Description string `bson:"description" json:"description"`

	Threads []any `bson:"threads,omitempty" json:"threads"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at" json:"deleted_at"`

	PostRef uint64 `bson:"post_ref" json:"post_ref"`
}

// Creates a new board with an ID and other default values.
func NewBoard() *Board {
	return &Board{
		ID:        primitive.NewObjectID(),
		Threads:   []any{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		PostRef:   0,
	}
}

// takes a bson.M, marshals it into bytes then the bytes into a Board struct
func UnmarshalBoard(d bson.M, t *Board) error {
	fmt.Println("Unmarshalling board")

	bs, err := bson.Marshal(d)
	if err != nil {
		fmt.Println("Error marshalling board:", err)
		return err
	}

	err = bson.Unmarshal(bs, t)
	if err != nil {
		fmt.Println("Error unmarshalling board:", err)
		return err
	}

	fmt.Println("Finished unmarshalling board")
	return nil
}
