package web

import "errors"

// shutdownError is a type to help with graceful termination of the service.
type shutdownError struct {
	Message string
}

// NewShutdownError returns an error that causes the framework to signal
// a graceful shutdown.
func NewShutdownError(message string) error {
	return &shutdownError{message}
}

// Error implements the error interface.
func (e *shutdownError) Error() string {
	return e.Message
}

// IsShutdownError checks to see if the shutdown error is contained
// in the specified error value.
func IsShutdownError(err error) bool {
	var e *shutdownError
	return errors.As(err, &e)
}
