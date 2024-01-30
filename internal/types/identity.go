package types

import (
	"math/rand"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// character sets
	IDENTITY_CHAR_SET = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789-_"

	// permissions
	PUBLIC_IDENTITY_FIELDS   = []string{"name", "style", "role", "status", "created_at", "updated_at", "deleted_at"}
	PERSONAL_IDENTITY_FIELDS = []string{"_id", "account"}

	// styles
	IDENTITY_STYLE_PREFIXES = []string{"filled", "ghost", "soft", "glass"}
	IDENTITY_STYLE_SUFFIXES = []string{"primary", "secondary", "tertiary", "success", "warning", "error", "surface"}
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
	name, _ := gonanoid.Generate(IDENTITY_CHAR_SET, 8)

	return &Identity{
		ID:        primitive.NewObjectID(),
		Role:      ThreadRoleUser,
		Status:    IdentityStatusActive,
		Style:     NewIdentityStyle(),
		Name:      name,
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

// ClientFormatter implementation
func (i *Identity) CLFormat() bson.M {
	return bson.M{
		"name":       i.Name,
		"style":      i.Style,
		"role":       i.Role,
		"status":     i.Status,
		"thread":     i.Thread,
		"created_at": i.CreatedAt,
		"updated_at": i.UpdatedAt,
	}
}

func randPrefix() string {
	return IDENTITY_STYLE_PREFIXES[rand.Intn(len(IDENTITY_STYLE_PREFIXES))]
	// return rand.Intn(max-min) + min
}

func randSuffix() string {
	return IDENTITY_STYLE_SUFFIXES[rand.Intn(len(IDENTITY_STYLE_SUFFIXES))]
}

func NewIdentityStyle() string {
	return "variant-" + randPrefix() + "-" + randSuffix()
}
