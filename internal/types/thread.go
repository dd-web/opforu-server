package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Thread struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	Title string `bson:"title" json:"title"`
	Body  string `bson:"body" json:"body"`

	Slug string `bson:"slug" json:"slug"`

	Board primitive.ObjectID `bson:"board" json:"board"`

	Account primitive.ObjectID `bson:"account" json:"account"`
	Creator primitive.ObjectID `bson:"creator" json:"creator"`

	Posts []primitive.ObjectID `bson:"posts" json:"posts"`
	Mods  []primitive.ObjectID `bson:"mods" json:"mods"`

	Status ThreadStatus `bson:"status" json:"status"`
	Tags   []string     `bson:"tags" json:"tags"`

	Flags []ThreadFlag `bson:"flags" json:"flags"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at" json:"deleted_at"`
}

type ThreadStatus string

const (
	ThreadStatusOpen    ThreadStatus = "open"
	ThreadStatusClosed  ThreadStatus = "closed"
	ThreadStatusPending ThreadStatus = "pending"
)

type ThreadRole string

const (
	ThreadRoleUser    ThreadRole = "user"
	ThreadRoleMod     ThreadRole = "mod"
	ThreadRoleCreator ThreadRole = "creator"
)

type ThreadFlag uint

const (
	ThreadFlagNone   ThreadFlag = iota
	ThreadFlagSticky ThreadFlag = 1 << iota
	ThreadFlagLocked ThreadFlag = 1 << iota
	ThreadFlagHidden ThreadFlag = 1 << iota
)

// takes a bson.M, marshals it into bytes then the bytes into a Thread struct
func UnmarshalThread(d bson.M, t *Thread) error {
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
