package auth

import (
	"errors"
	"fmt"
)

type Error struct {
	msg string
}

func NewError(format string, args ...any) error {
	return &Error{
		msg: fmt.Sprintf(format, args...),
	}
}

func (ae *Error) Error() string {
	return ae.msg
}

func IsAuthError(err error) bool {
	var ae *Error
	return errors.As(err, &ae)
}
