package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// permissions
	PUBLIC_BOARD_FIELDS = []string{"title", "short", "description", "threads", "created_at", "updated_at", "deleted_at"}
)

type Board struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	Title       string `bson:"title" json:"title"`
	Short       string `bson:"short" json:"short"` // short name for the board (used in URLs)
	Description string `bson:"description" json:"description"`

	Threads []primitive.ObjectID `bson:"threads,omitempty" json:"threads,omitempty"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`

	PostRef uint64 `bson:"post_ref" json:"post_ref"`
}

// Creates a new board with an ID and other default values.
func NewBoard() *Board {
	ts := time.Now().UTC()
	return &Board{
		ID:        primitive.NewObjectID(),
		Threads:   []primitive.ObjectID{},
		CreatedAt: &ts,
		UpdatedAt: &ts,
		PostRef:   0,
	}
}
