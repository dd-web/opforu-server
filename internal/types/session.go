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
	PERSONAL_SESSION_FIELDS = []string{"_id", "account_id", "session_id", "expires"}
)

type Session struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`

	SessionID string             `bson:"session_id" json:"session_id"`
	AccountID primitive.ObjectID `bson:"account_id" json:"account_id"`
	Account   *Account           `bson:"account" json:"account"`

	Expires *time.Time `bson:"expires" json:"expires"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// creates a new session for the given account
func NewSession(account *Account) *Session {
	id := uuid.NewString()
	ts := time.Now().UTC()
	exp := time.Now().Add(time.Duration(time.Hour * 24 * 7)).UTC() // 1 week vailidity

	return &Session{
		ID:        primitive.NewObjectID(),
		SessionID: id,
		AccountID: account.ID,
		Account:   account,
		Expires:   &exp,
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

// is the session expired?
func (s *Session) IsExpired() bool {
	return s.Expires.Before(time.Now().UTC())
}

// will the session expire within the next hour?
func (s *Session) IsExpiringSoon() bool {
	return s.Expires.Before(time.Now().Add(-time.Duration(time.Minute * 60 * 24))) // 1 day refresh
}

// is the session expiring in the next 5 minutes?
func (s *Session) IsExpiryImminent() bool {
	return s.Expires.Before(time.Now().Add(-time.Duration(time.Minute * 5)))
}

// implements the ClientFormatter interface
func (s *Session) CLFormat() bson.M {
	return bson.M{
		"session_id": s.SessionID,
		"account_id": s.AccountID,
		"expires":    s.Expires,
	}
}
