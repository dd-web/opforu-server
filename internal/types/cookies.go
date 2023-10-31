package types

import (
	"net/http"
	"time"
)

// func NewSessionCookie() *http.Cookie {
// 	return &http.Cookie{
// 		Name:     "session_id",
// 		Value:    uuid.NewString(),
// 		Path:     "/",
// 		Expires:  time.Now().Add(time.Duration(SECONDS_IN_DAY) * time.Second).UTC(),
// 		MaxAge:   SECONDS_IN_DAY,
// 		Secure:   true,
// 		HttpOnly: true,
// 		SameSite: http.SameSiteStrictMode,
// 	}
// }

// helper method for deleting cookies on the client side
func NewCookieDeleter() *http.Cookie {
	return &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
