package types

import (
	"net/http"
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

	MaxAge   int           `bson:"max_age" json:"max_age"`
	HttpOnly bool          `bson:"http_only" json:"http_only"`
	Secure   bool          `bson:"secure" json:"secure"`
	Path     string        `bson:"path" json:"path"`
	SameSite http.SameSite `bson:"same_site" json:"same_site"`
	Domain   string        `bson:"domain" json:"domain"`

	Expires *time.Time `bson:"expires" json:"expires"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// creates a new session for the given account
func NewSession(account primitive.ObjectID) *Session {
	id := uuid.NewString()
	ts := time.Now().UTC()
	exp := time.Now().Add(time.Duration(SECONDS_IN_DAY) * time.Second).UTC()

	return &Session{
		ID:        primitive.NewObjectID(),
		SessionID: id,
		AccountID: account,
		MaxAge:    SECONDS_IN_DAY,
		HttpOnly:  true,
		Secure:    true,
		Path:      "/",
		SameSite:  http.SameSiteStrictMode,
		Expires:   &exp,
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

// formats session into an http cookie
func (s *Session) CookieFromSession() *http.Cookie {
	return &http.Cookie{
		Name:     "session_id",
		Value:    s.SessionID,
		Expires:  *s.Expires,
		MaxAge:   s.MaxAge,
		HttpOnly: s.HttpOnly,
		Secure:   s.Secure,
		Path:     s.Path,
		SameSite: s.SameSite,
		Domain:   s.Domain,
	}
}

// is the session expired?
func (s *Session) IsExpired() bool {
	return s.Expires.Before(time.Now().UTC())
}

// format a session for the client (set into cookie from sveltekit)
func (s *Session) FormatForClient() bson.M {
	return bson.M{
		"session_id": s.SessionID,
		"expiry":     s.Expires,
	}
}

// for unmarshalling a session from the request into the session object to be passes down the chain
// func UnmarshalSession(rc *RequestCtx) {
// 	var err error

// 	body, err := io.ReadAll(rc.Request.Body)
// 	if err != nil {
// 		fmt.Println("error reading body")
// 		return
// 	}

// 	var parsed struct {
// 		Session string `json:"session"`
// 	}

// 	err = json.Unmarshal(body, &parsed)
// 	if err != nil {
// 		fmt.Println("error unmarshalling body")
// 		return
// 	}

// 	foundSession, err := rc.Store.FindSession(parsed.Session)
// 	if err != nil {
// 		return
// 	}

// 	if foundSession.IsExpired() {
// 		fmt.Println("session expired")
// 		return
// 	}

// 	account, err := rc.Store.FindAccountByID(foundSession.AccountID)
// 	if err != nil {
// 		fmt.Println("could not find the user account")
// 		return
// 	}

// 	rc.AccountCtx.Account = account
// 	rc.AccountCtx.Session = foundSession

// }

// func ResolveSessionFromCtx(rc *RequestCtx) *Session {
// 	return NewSession(rc.AccountCtx.Account.ID)
// }
