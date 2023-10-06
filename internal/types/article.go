package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Article struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	AuthorID  primitive.ObjectID   `bson:"author" json:"author"`
	CoAuthors []primitive.ObjectID `bson:"co_authors" json:"co_authors"`

	// user content fields
	Title string
	Body  string
	Slug  string
	Tags  []string

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at" json:"deleted_at"`
}

func NewArticle() *Article {
	return &Article{
		ID:        primitive.NewObjectID(),
		CoAuthors: []primitive.ObjectID{},
		Tags:      []string{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// takes a bson.M, marshals it into bytes then the bytes into an Article struct
func UnmarshalArticle(d bson.M, t *Article) error {
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
