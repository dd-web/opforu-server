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

	Board   primitive.ObjectID `bson:"board" json:"board"`
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
		Slug:      NewThreadSlug(),
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

// Bitfield flags for threads
type ThreadFlag uint

const (
	ThreadFlagNone   ThreadFlag = iota      // thread has no flags
	ThreadFlagSticky ThreadFlag = 1 << iota // thread is sticky
	ThreadFlagLocked ThreadFlag = 1 << iota // thread is locked
	ThreadFlagHidden ThreadFlag = 1 << iota // thread is hidden
	ThreadFlagNSFW   ThreadFlag = 1 << iota // thread is NSFW
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

func (t *Thread) Validate() error {
	if len(t.Title) < 5 {
		return fmt.Errorf("Thread title is too short")
	} else if len(t.Title) > 120 {
		return fmt.Errorf("Thread title is too long")
	}

	if len(t.Body) < 5 {
		return fmt.Errorf("Thread body is too short")
	} else if len(t.Body) > 4200 {
		return fmt.Errorf("Thread body is too long")
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
