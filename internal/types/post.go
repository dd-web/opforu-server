package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// permissions
	PUBLIC_POST_FIELDS = []string{"post_number", "body", "assets", "creator", "board", "thread", "created_at", "updated_at", "deleted_at"}
	ADMIN_POST_FIELDS  = []string{"_id", "account"}
)

type Post struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	PostNumber uint64             `bson:"post_number" json:"post_number"`
	Creator    primitive.ObjectID `bson:"creator" json:"creator"` // identity _id

	Body   string               `bson:"body" json:"body"`
	Assets []primitive.ObjectID `bson:"assets" json:"assets"`

	Board  primitive.ObjectID `bson:"board" json:"board"`
	Thread primitive.ObjectID `bson:"thread" json:"thread"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// Creates a new post with an ID and other default values.
func NewPost() *Post {
	ts := time.Now().UTC()
	return &Post{
		ID:        primitive.NewObjectID(),
		Assets:    []primitive.ObjectID{},
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

// ClientFormatter implementation
func (p *Post) CLFormat() bson.M {
	return bson.M{
		"post_number": p.PostNumber,
		"body":        p.Body,
		"assets":      p.Assets,
		"creator":     p.Creator,
		"board":       p.Board,
		"thread":      p.Thread,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	}
}
