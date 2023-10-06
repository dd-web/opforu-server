package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Board struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	Title       string `bson:"title" json:"title"`
	Short       string `bson:"short" json:"short"` // short name for the board (used in URLs)
	Description string `bson:"description" json:"description"`

	Threads []primitive.ObjectID `bson:"threads" json:"threads"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at" json:"deleted_at"`

	PostRef uint64 `bson:"post_ref" json:"post_ref"`
}

// Creates a new board with an ID and other default values.
func NewBoard() *Board {
	return &Board{
		ID:        primitive.NewObjectID(),
		Threads:   []primitive.ObjectID{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		PostRef:   0,
	}
}
