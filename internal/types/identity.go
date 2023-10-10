package types

import (
	"time"

	"github.com/dd-web/opforu-server/internal/utils"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Identity struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Account primitive.ObjectID `bson:"account,omitempty" json:"account"`

	Name  string `bson:"name" json:"name"`
	Style string `bson:"style" json:"style"`

	Role   ThreadRole     `bson:"role" json:"role"`
	Status IdentityStatus `bson:"status" json:"status"`

	Thread primitive.ObjectID `bson:"thread,omitempty" json:"thread"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type IdentityStatus string

const (
	IdentityStatusUnknown   IdentityStatus = "unknown"
	IdentityStatusActive    IdentityStatus = "active"
	IdentityStatusSuspended IdentityStatus = "suspended"
	IdentityStatusBanned    IdentityStatus = "banned"
	IdentityStatusDeleted   IdentityStatus = "deleted"
)

// Creates a new Identity object with some values set by default
func NewIdentity() *Identity {
	ts := time.Now().UTC()
	name, _ := gonanoid.Generate(utils.GetIdentityCharSet(), 8)

	return &Identity{
		ID:        primitive.NewObjectID(),
		Role:      ThreadRoleUser,
		Status:    IdentityStatusActive,
		Name:      name,
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

// takes a bson.M, marshals it into bytes then the bytes into a Identity struct
func UnmarshalIdentity(d bson.M, t *Identity) error {
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
