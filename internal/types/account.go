package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Account struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username string             `bson:"username,omitempty" json:"username"`
	Email    string             `bson:"email,omitempty" json:"email"`

	Role   AccountRole   `bson:"role" json:"role"`
	Status AccountStatus `bson:"status" json:"status"`

	Password string `json:"password_hash" bson:"password_hash"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type AccountStatus string

const (
	AccountStatusUnknown   AccountStatus = "unknown"
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusBanned    AccountStatus = "banned"
	AccountStatusDeleted   AccountStatus = "deleted"
)

type AccountRole string

const (
	AccountRoleUnknown AccountRole = "unknown"
	AccountRolePublic  AccountRole = "public"
	AccountRoleUser    AccountRole = "user"
	AccountRoleMod     AccountRole = "mod"
	AccountRoleAdmin   AccountRole = "admin"
)

// Creates a new Account object with some values set by default
func NewAccount() *Account {
	ts := time.Now().UTC()
	return &Account{
		ID:        primitive.NewObjectID(),
		Role:      AccountRoleUser,
		Status:    AccountStatusActive,
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

// ClientFormatter implementation
func (a *Account) CLFormat() bson.M {
	return bson.M{
		"username":   a.Username,
		"email":      a.Email,
		"role":       a.Role,
		"status":     a.Status,
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}
}

func (a *Account) IsAdmin() bool {
	return a.Role == AccountRoleAdmin
}

func (a *Account) IsMod() bool {
	return a.Role == AccountRoleMod
}

func (a *Account) IsStaff() bool {
	return a.IsAdmin() || a.IsMod()
}
