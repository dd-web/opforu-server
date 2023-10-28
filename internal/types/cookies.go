package types

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type CookieVariant string

const (
	COOKIE_SESSION CookieVariant = "session"
)

func NewSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "session_id",
		Value:    uuid.NewString(),
		Expires:  time.Now().Add(time.Duration(SECONDS_IN_DAY) * time.Second).UTC(),
		MaxAge:   SECONDS_IN_DAY,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
