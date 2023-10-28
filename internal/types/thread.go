package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// permissions
	PUBLIC_THREAD_FIELDS = []string{"title", "body", "slug", "board", "creator", "posts", "mods", "status", "tags", "created_at", "updated_at", "deleted_at"}
	MOD_THREAD_FIELDS    = []string{"flags"}
	ADMIN_THREAD_FIELDS  = []string{"_id", "account"}

	// character sets
	THREAD_SLUG_CHAR_SET = "abcdefghijklmnopqrstuvwxyz0123456789"
)

type Thread struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	Title string `bson:"title" json:"title"`
	Body  string `bson:"body" json:"body"`
	Slug  string `bson:"slug" json:"slug"`

	Board primitive.ObjectID `bson:"board" json:"board"`

	Account primitive.ObjectID `bson:"account" json:"account"`
	Creator primitive.ObjectID `bson:"creator" json:"creator"`

	Posts []primitive.ObjectID `bson:"posts" json:"posts"`
	Mods  []primitive.ObjectID `bson:"mods" json:"mods"`

	Status ThreadStatus `bson:"status" json:"status"`
	Tags   []string     `bson:"tags" json:"tags"`

	Flags []ThreadFlag `bson:"flags" json:"flags"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
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
