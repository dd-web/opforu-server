package types

import "net/http"

type APIError struct {
	Status  int
	Message string `json:"error"`
}

func (e APIError) Error() string {
	return e.Message
}

func ErrorNotFound(s string) APIError {
	return APIError{
		Status:  http.StatusNotFound,
		Message: s + "not found",
	}
}
