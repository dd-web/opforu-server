package types

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

	CookieType CookieVariant `bson:"cookie_type" json:"cookie_type"`

	Flags []SessionFlag `bson:"flags" json:"flags"`

	Cookie *http.Cookie `bson:"-" json:"-"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func NewSession(userId primitive.ObjectID) *Session {
	cookie := NewSessionCookie()
	ts := time.Now().UTC()
	return &Session{
		ID:         primitive.NewObjectID(),
		SessionID:  cookie.Value,
		AccountID:  userId,
		CookieType: COOKIE_SESSION,
		Cookie:     cookie,
		CreatedAt:  &ts,
		UpdatedAt:  &ts,
	}

	// now := time.Now().UTC()
	// expires := time.Now().Add(time.Duration(SECONDS_IN_DAY) * time.Second).UTC() // 24 hours from now
	// id := uuid.New().String()
	// return &Session{
	// 	ID:        primitive.NewObjectID(),
	// 	AccountID: userId,
	// 	SessionID: id,
	// 	Active:    true,
	// 	Expiry:    &expires,
	// 	CreatedAt: &now,
	// 	UpdatedAt: &now,
	// }
}

// is the session expired?
func (s *Session) IsExpired() bool {
	return s.Cookie.Expires.Before(time.Now().UTC())
}

// format a session for the client (set into cookie from sveltekit)
func (s *Session) FormatForClient() bson.M {
	return bson.M{
		"session_id": s.SessionID,
		"expiry":     s.Cookie.Expires,
	}
}

// for unmarshalling a session from the request into the session object to be passes down the chain
func UnmarshalSession(rc *RequestCtx) {
	var err error

	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		fmt.Println("error reading body")
		return
	}

	var parsed struct {
		Session string `json:"session"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println("error unmarshalling body")
		return
	}

	foundSession, err := rc.Store.FindSession(parsed.Session)
	if err != nil {
		return
	}

	if foundSession.IsExpired() {
		fmt.Println("session expired")
		return
	}

	account, err := rc.Store.FindAccountByID(foundSession.AccountID)
	if err != nil {
		fmt.Println("could not find the user account")
		return
	}

	rc.AccountCtx.Account = account
	rc.AccountCtx.Session = foundSession

}

func ResolveSessionFromCtx(rc *RequestCtx) *Session {
	return NewSession(rc.AccountCtx.Account.ID)
}

// Bitfield flags for sessions
type SessionFlag uint

const (
	SessionFlagNone              SessionFlag = iota      // ssession has no flags
	SessionFlagMarkedForDeletion SessionFlag = 1 << iota // session is marked for deletion
)
