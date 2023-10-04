package request

import (
	"errors"
)

// ErrorResponse represents error response.
type ErrorResponse struct {
	Error  string            `json:"message"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Error is used to pass an error during the request through the
// application with web specific context. is the form used for API responses from failures in the API.
type Error struct {
	Err    error
	Status int
}

// NewError wraps a provided error with an HTTP status code.
// This function should ve used when handlers encounter expected errors.
func NewError(err error, status int) error {
	return &Error{err, status}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the service's logs.
func (e *Error) Error() string {
	return e.Err.Error()
}

// IsRequestError checks if an error of type RequestError exists.
func IsError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

// GetError returns a copy of the RequestError pointer.
func GetError(err error) *Error {
	var e *Error
	if !errors.As(err, &e) {
		return nil
	}
	return e
}
