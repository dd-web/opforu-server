package types

import (
	"fmt"
	"math/rand"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// permissions
	PUBLIC_THREAD_FIELDS = []string{"title", "body", "slug", "board", "creator", "posts", "mods", "status", "tags", "created_at", "updated_at", "deleted_at"}
	MOD_THREAD_FIELDS    = []string{"flags"}
	ADMIN_THREAD_FIELDS  = []string{"_id", "account"}

	// character sets
	THREAD_SLUG_CHAR_SET = "abcdefghijklmnopqrstuvwxyz0123456789"
	THREAD_MIN_SLUG_LEN  = 8
	THREAD_MAX_SLUG_LEN  = 12
)

type Thread struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Status ThreadStatus       `bson:"status" json:"status"`

	Title string `bson:"title" json:"title"`
	Body  string `bson:"body" json:"body"`
	Slug  string `bson:"slug" json:"slug"`

	Board primitive.ObjectID `bson:"board" json:"board"`

	// Creator is the Identity made for the creator of the thread, not the account id
	Creator primitive.ObjectID `bson:"creator" json:"creator"`

	Posts []primitive.ObjectID `bson:"posts" json:"posts"`
	Mods  []primitive.ObjectID `bson:"mods" json:"mods"`

	Assets []primitive.ObjectID `bson:"assets" json:"assets"`

	Tags  []string     `bson:"tags" json:"tags"`
	Flags []ThreadFlag `bson:"flags" json:"flags"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func NewThread() *Thread {
	ts := time.Now().UTC()
	return &Thread{
		ID:        primitive.NewObjectID(),
		Status:    ThreadStatusOpen,
		Title:     "",
		Body:      "",
		Slug:      NewThreadSlug(),
		Board:     primitive.NilObjectID,
		Creator:   primitive.NilObjectID,
		Posts:     []primitive.ObjectID{},
		Mods:      []primitive.ObjectID{},
		Assets:    []primitive.ObjectID{},
		Tags:      []string{},
		Flags:     []ThreadFlag{},
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

type ThreadStatus string

const (
	ThreadStatusOpen     ThreadStatus = "open"
	ThreadStatusClosed   ThreadStatus = "closed"
	ThreadStatusArchived ThreadStatus = "archived"
	ThreadStatusDeleted  ThreadStatus = "deleted"
)

type ThreadRole string

const (
	ThreadRoleUser    ThreadRole = "user"
	ThreadRoleMod     ThreadRole = "mod"
	ThreadRoleCreator ThreadRole = "creator"
)

// user settable flags
type ThreadFlag string

const (
	TF_NSFW     ThreadFlag = "nsfw"
	TF_NSFL     ThreadFlag = "nsfl"
	TF_MEDIAREQ ThreadFlag = "media_required"
)

// ClientFormatter implementation
func (t *Thread) CLFormat() bson.M {
	return bson.M{
		"title":      t.Title,
		"body":       t.Body,
		"slug":       t.Slug,
		"board":      t.Board,
		"creator":    t.Creator,
		"posts":      t.Posts,
		"mods":       t.Mods,
		"status":     t.Status,
		"tags":       t.Tags,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
}

// for now use strings - eventually use bitfields for performance
func (t *Thread) AttachFlags(rft *RUMThreadFlags) {
	if rft.NSFW {
		t.Flags = append(t.Flags, TF_NSFW)
	}
	if rft.NSFL {
		t.Flags = append(t.Flags, TF_NSFL)
	}
	if rft.REQMEDIA {
		t.Flags = append(t.Flags, TF_MEDIAREQ)
	}
}

func (t *Thread) HasFlag(flag ThreadFlag) bool {
	for _, v := range t.Flags {
		if v == flag {
			return true
		}
	}
	return false
}

func (t *Thread) Validate() error {
	if len(t.Title) < 2 {
		return fmt.Errorf("Thread title is too short")
	} else if len(t.Title) > 120 {
		return fmt.Errorf("Thread title is too long")
	}

	if len(t.Slug) < THREAD_MIN_SLUG_LEN {
		return fmt.Errorf("Thread slug is too short")
	} else if len(t.Slug) > THREAD_MAX_SLUG_LEN {
		return fmt.Errorf("Thread slug is too long")
	}

	if len(t.Tags) > 5 {
		return fmt.Errorf("Thread has too many tags")
	}

	if len(t.Assets) > 9 {
		return fmt.Errorf("Thread has too many assets")
	}

	return nil
}

func NewThreadSlug() string {
	slugLen := rand.Intn(THREAD_MAX_SLUG_LEN-THREAD_MIN_SLUG_LEN) + THREAD_MIN_SLUG_LEN
	slug, _ := gonanoid.Generate(THREAD_SLUG_CHAR_SET, slugLen)
	return slug
}
