package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// permissions
	PUBLIC_ARTICLE_FIELDS = []string{"title", "body", "slug", "author", "co_authors", "status", "tags", "created_at", "updated_at", "deleted_at"}
)

type Article struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	AuthorID  primitive.ObjectID   `bson:"author" json:"author"`
	CoAuthors []primitive.ObjectID `bson:"co_authors" json:"co_authors"`

	Status ArticleStatus `bson:"status" json:"status"`

	Title string   `bson:"title" json:"title"`
	Body  string   `bson:"body" json:"body"`
	Slug  string   `bson:"slug" json:"slug"`
	Tags  []string `bson:"tags" json:"tags"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"
	ArticleStatusDeleted   ArticleStatus = "deleted"
)

func NewArticle() *Article {
	ts := time.Now().UTC()
	return &Article{
		ID:        primitive.NewObjectID(),
		CoAuthors: []primitive.ObjectID{},
		Tags:      []string{},
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}
