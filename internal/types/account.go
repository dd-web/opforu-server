package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Account struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username string             `bson:"username,omitempty" json:"username"`
	Email    string             `bson:"email,omitempty" json:"email"`

	Role   AccountRole   `bson:"role" json:"role"`
	Status AccountStatus `bson:"status" json:"status"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at" json:"deleted_at"`
}

type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusBanned    AccountStatus = "banned"
	AccountStatusDeleted   AccountStatus = "deleted"
)

type AccountRole string

const (
	AccountRolePublic AccountRole = "public"
	AccountRoleUser   AccountRole = "user"
	AccountRoleMod    AccountRole = "mod"
	AccountRoleAdmin  AccountRole = "admin"
)

// Creates a new Account object with some values set by default
func NewAccount() *Account {
	return &Account{
		ID:        primitive.NewObjectID(),
		Role:      AccountRoleUser,
		Status:    AccountStatusActive,
		CreatedAt: time.Now().UTC(),
	}
}
