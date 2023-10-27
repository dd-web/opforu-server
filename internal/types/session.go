package types

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// measurements of time
	SECONDS_IN_MINUTE = 60
	MINUTES_IN_HOUR   = 60
	HOURS_IN_DAY      = 24

	DAYS_IN_WEEK = 7
	DAYS_IN_YEAR = 365

	SECONDS_IN_HOUR = SECONDS_IN_MINUTE * MINUTES_IN_HOUR
	SECONDS_IN_DAY  = SECONDS_IN_HOUR * HOURS_IN_DAY
	SECONDS_IN_WEEK = SECONDS_IN_DAY * DAYS_IN_WEEK

	MINUTES_IN_DAY  = MINUTES_IN_HOUR * HOURS_IN_DAY
	MINUTES_IN_WEEK = MINUTES_IN_DAY * DAYS_IN_WEEK

	HOURS_IN_WEEK = HOURS_IN_DAY * DAYS_IN_WEEK
	HOURS_IN_YEAR = HOURS_IN_DAY * DAYS_IN_YEAR

	// permissions
	PUBLIC_SESSION_FIELDS   = []string{"created_at", "updated_at", "deleted_at"}
	PERSONAL_SESSION_FIELDS = []string{"_id", "account_id", "session_id", "active", "expiry"}
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

const ()

func NewSession(userId primitive.ObjectID) *Session {
	now := time.Now().UTC()
	expires := time.Now().Add(time.Duration(SECONDS_IN_DAY) * time.Second).UTC() // 24 hours from now
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
