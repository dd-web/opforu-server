package types

import "net/http"

// This should be the ONLY error type a client receives. Ever. Read top of file for more info.
type APIError struct {
	Status  int
	Message string `json:"error"`
}

// implements the error interface
func (e APIError) Error() string {
	return e.Message
}

// creates a new APIError and returns a pointer to it. you should return the memory address of this
// when returning an error from a handler. it's a pointer here because it doesn't need to be copied
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

// Error Not Found
// happens when the server can't find the requested resources. this is a 404 error.
// try making your query less specific or check the spelling of the resource you're requesting.
// takes a string name of the resource
func ErrorNotFound(s string) APIError {
	return *NewAPIError(http.StatusNotFound, "not found: "+s)
}
