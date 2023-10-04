package request

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewError(t *testing.T) {
	e := errors.New("msg")
	assert.Equal(t, NewError(e, 400), &Error{e, 400})
	assert.Equal(t, NewError(e, 400).Error(), e.Error())
}

func Test_IsError(t *testing.T) {
	e := errors.New("msg")
	assert.False(t, IsError(e))
	assert.True(t, IsError(NewError(e, 400)))
}

func Test_GetError(t *testing.T) {
	e := errors.New("msg")
	requestErr := NewError(e, 400)
	assert.Nil(t, GetError(e))
	assert.Equal(t, requestErr.Error(), GetError(requestErr).Error())
}
