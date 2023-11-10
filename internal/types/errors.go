package types

import (
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type ServerError string

// given strings for each error type
const (
	Error_Unexpected   ServerError = "unexpected server error"
	Error_NotFound     ServerError = "not found"
	Error_Invalid      ServerError = "invalid"
	Error_Unauthorized ServerError = "unauthorized"
)

// Status codes mapped to their respective ServerError
var ServerErrorStatusMap map[int]ServerError = map[int]ServerError{
	http.StatusInternalServerError: Error_Unexpected,
	http.StatusNotFound:            Error_NotFound,
	http.StatusBadRequest:          Error_Invalid,
	http.StatusUnauthorized:        Error_Unauthorized,
}

func (se ServerError) String() string {
	return string(se)
}

// This should be the ONLY error type a client receives. Ever. Read top of file for more info.
type APIError struct {
	Status  int
	Message string
}

// implements the error interface
func (e APIError) Error() string {
	return e.Message
}

// formats into bson (json likeish) for response
func (e APIError) Bson() bson.M {
	return bson.M{"error": e.Message}
}

// Creates a new APIError with the given status and message
func NewAPIError(status int, message string) *APIError {
	return &APIError{
		Status:  status,
		Message: message,
	}
}

/***********************************************************************************************/
/* CONSTRUCTORS - below this line should only be funcs that return new APIErrors. Please write
/*  comments explaining what the error is, how to avoid it, and any arguments it takes.
/***********************************************************************************************/

// New Not Found Error
// - accepts a string of the resource that was not found
func ErrorNotFound(resource string) APIError {
	return *NewAPIError(http.StatusNotFound, resource+" "+Error_NotFound.String())
}

// New Invalid Error
// - accepts a string of what was invalid
func ErrorInvalid(what string) APIError {
	return *NewAPIError(http.StatusBadRequest, what+" is "+Error_Invalid.String())
}

// New Conflict Error
func ErrorConflict(msg string) APIError {
	return *NewAPIError(http.StatusConflict, msg)
}

// New Unauthorized Error
func ErrorUnauthorized() APIError {
	return *NewAPIError(http.StatusUnauthorized, Error_Unauthorized.String())
}

// New Unexpected Error
func ErrorUnexpected() APIError {
	return *NewAPIError(http.StatusInternalServerError, Error_Unexpected.String())
}
