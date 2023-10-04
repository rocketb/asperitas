package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespond(t *testing.T) {
	type args struct {
		data       any
		statusCode int
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "respond with json data",
			args: args{
				statusCode: http.StatusOK,
				data:       "",
			},
			wantErr: assert.NoError,
		},
		{
			name: "no content",
			args: args{
				statusCode: http.StatusNoContent,
			},
			wantErr: assert.NoError,
		},
		{
			name: "err marshalling",
			args: args{
				data:       make(chan int),
				statusCode: http.StatusOK,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := Respond(context.Background(), w, tt.args.data, tt.args.statusCode)
			tt.wantErr(t, err, fmt.Sprintf("Respond(r, %v, %v)", tt.args.data, tt.args.statusCode))

			if err != nil || tt.args.statusCode == http.StatusNoContent {
				return
			}

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, fmt.Sprintf("\"%s\"", tt.args.data), w.Body.String())
		})
	}
}
