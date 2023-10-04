package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_wrapMiddleware(t *testing.T) {
	type args struct {
		mw      []Middleware
		handler Handler
	}
	tests := []struct {
		name string
		args args
		want Handler
	}{
		{
			name: "empty mw",
			args: args{
				mw: []Middleware{},
				handler: func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return nil
				},
			},
			want: func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return nil
			},
		},
		{
			name: "mw wrapped correctly",
			args: args{
				mw: []Middleware{
					func(handler Handler) Handler {
						return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
							return nil
						}
					},
				},
				handler: func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					return errors.New("error")
				},
			},
			want: func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := wrapMiddleware(tt.args.mw, tt.args.handler)
			w := httptest.NewRecorder()
			assert.Equal(t, tt.want(context.Background(), w, &http.Request{}), wr(context.Background(), w, &http.Request{}))
		})
	}
}
