package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	type Dump struct {
		Field string `json:"field" validate:"required,len=8"`
	}
	type args struct {
		body any
		val  any
	}
	tests := []struct {
		name       string
		args       args
		wantErr    assert.ErrorAssertionFunc
		wantErrMsg string
	}{
		{
			name: "success",
			args: args{
				body: struct {
					Field string `json:"field"`
				}{
					Field: "12345678",
				},
				val: &Dump{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "err on decode",
			args: args{
				body: struct {
					OtherField string `json:"other_field"`
				}{
					OtherField: "12345678",
				},
				val: &Dump{},
			},
			wantErr:    assert.Error,
			wantErrMsg: errors.New("unable to decode payload: json: unknown field \"other_field\"").Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.args.body)
			r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(b))
			err := Decode(r, tt.args.val)
			if !tt.wantErr(t, err) {
				assert.EqualError(t, err, tt.wantErrMsg)
			}
		})
	}
}
