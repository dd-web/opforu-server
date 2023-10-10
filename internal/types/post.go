package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	PostNumber uint64 `bson:"post_number" json:"post_number"`

	Body string `bson:"body" json:"body"`

	Assets []primitive.ObjectID `bson:"assets" json:"assets"`

	Account primitive.ObjectID `bson:"account" json:"account"`
	Creator primitive.ObjectID `bson:"creator" json:"creator"`

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

// takes a bson.M, marshals it into bytes then the bytes into a Post struct
func UnmarshalPost(d bson.M, t *Post) error {
	bs, err := bson.Marshal(d)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(bs, t)
	if err != nil {
		return err
	}
	return nil
}
