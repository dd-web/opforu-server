package types

import (
	"math/rand"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	PUBLIC_ARTICLE_FIELDS = []string{"title", "body", "slug", "author", "co_authors", "status", "tags", "created_at", "updated_at", "deleted_at"}

	// char sets
	// min/max only apply to randomly generated slugs. reserved/specified slugs can be up to 128 characters.
	ARTICLE_SLUG_CHAR_SET = "abcdefghijklmnopqrstuvwxyz0123456789"
	ARTICLE_MAX_SLUG_LEN  = 16
	ARTICLE_MIN_SLUG_LEN  = 10
)

type Article struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`

	AuthorID  primitive.ObjectID   `json:"author" bson:"author"`         // ArticleAuthor id
	CoAuthors []primitive.ObjectID `json:"co_authors" bson:"co_authors"` // ArticleAuthor id's

	Status     ArticleStatus `bson:"status" json:"status"`
	CommentRef int           `json:"comment_ref" bson:"comment_ref"`

	Comments []primitive.ObjectID `json:"comments" bson:"comments"`
	Assets   []primitive.ObjectID `json:"assets" bson:"assets"`

	Title string   `json:"title" bson:"title"`
	Body  string   `json:"body" bson:"body"`
	Slug  string   `json:"slug" bson:"slug"`
	Tags  []string `json:"tags" bson:"tags"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func NewArticle() *Article {
	ts := time.Now().UTC()
	return &Article{
		ID:         primitive.NewObjectID(),
		AuthorID:   primitive.NilObjectID,
		CoAuthors:  []primitive.ObjectID{},
		Status:     ArticleStatusDraft,
		CommentRef: 0,
		Comments:   []primitive.ObjectID{},
		Assets:     []primitive.ObjectID{},
		Title:      "",
		Body:       "",
		Slug:       NewArticleSlug(),
		Tags:       []string{},
		CreatedAt:  &ts,
		UpdatedAt:  &ts,
	}
}

type ArticleComment struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	AuthorID   primitive.ObjectID `json:"author" bson:"author"`                     // account id
	AuthorAnon bool               `json:"author_anonymous" bson:"author_anonymous"` // only admins/mods have the option

	CommentNumber int                  `json:"comment_number" bson:"comment_number"`
	Body          string               `json:"body" bson:"body"`
	Assets        []primitive.ObjectID `json:"assets" bson:"assets"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func NewArticleComment() *ArticleComment {
	ts := time.Now().UTC()
	return &ArticleComment{
		ID:            primitive.NewObjectID(),
		AuthorID:      primitive.NilObjectID,
		AuthorAnon:    false,
		CommentNumber: 0,
		Body:          "",
		Assets:        []primitive.ObjectID{},
		CreatedAt:     &ts,
		UpdatedAt:     &ts,
	}
}

type ArticleAuthor struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	AuthorID  primitive.ObjectID `json:"author" bson:"author"`       // account id
	Anonymize bool               `json:"anonymize" bson:"anonymize"` // make author/coauthor with above id anonymous
}

func NewArticleAuthor() *ArticleAuthor {
	return &ArticleAuthor{
		ID:        primitive.NewObjectID(),
		AuthorID:  primitive.NilObjectID,
		Anonymize: false,
	}
}

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"
	ArticleStatusDeleted   ArticleStatus = "deleted"
)

// ClientFormatter implementation
func (a *Article) CLFormat() bson.M {
	return bson.M{
		"title":      a.Title,
		"body":       a.Body,
		"slug":       a.Slug,
		"author":     a.AuthorID,
		"co_authors": a.CoAuthors,
		"status":     a.Status,
		"tags":       a.Tags,
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}
}

func NewArticleSlug() string {
	slugLen := rand.Intn(ARTICLE_MAX_SLUG_LEN-ARTICLE_MIN_SLUG_LEN) + ARTICLE_MAX_SLUG_LEN
	slug, _ := gonanoid.Generate(ARTICLE_SLUG_CHAR_SET, slugLen)
	return slug
}
