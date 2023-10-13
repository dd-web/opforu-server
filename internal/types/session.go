package types

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	SessionID string             `bson:"session_id" json:"session_id"`
	AccountID primitive.ObjectID `bson:"account_id" json:"account_id"`

	Active bool `bson:"active" json:"active"`

	Expiry    *time.Time `bson:"expiry" json:"expiry"`
	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

const (
	SECONDS_IN_HOUR = 3600
	HOURS_IN_DAY    = 24
	SECOND_IN_DAY   = SECONDS_IN_HOUR * HOURS_IN_DAY
)

func NewSession(userId primitive.ObjectID) *Session {
	now := time.Now().UTC()
	expires := time.Now().Add(SECOND_IN_DAY * time.Second).UTC() // 24 hours from now
	id := uuid.New().String()
	return &Session{
		ID:        primitive.NewObjectID(),
		AccountID: userId,
		SessionID: id,
		Active:    true,
		Expiry:    &expires,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
}

// is the session expired?
func (s *Session) IsExpired() bool {
	return s.Expiry.Before(time.Now().UTC())
}

// format a session for the client (set into cookie from sveltekit)
func (s *Session) FormatForClient() bson.M {
	return bson.M{
		"session_id": s.SessionID,
		"expiry":     s.Expiry,
	}
}
