package postgrp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/post"
	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/internal/web/paging"

	"github.com/rocketb/asperitas/pkg/web"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	curDate = time.Time{}
	tAuthor = user.User{
		ID:   uuid.New(),
		Name: "User1",
	}
	tVotes       = []post.Vote{}
	tComments    = []post.Comment{}
	tCommAuthors = map[uuid.UUID]user.User{}
	tPost        = post.Post{
		ID:          uuid.New(),
		Type:        "text",
		Title:       "title",
		Body:        "text",
		Category:    "category",
		Score:       1,
		Views:       1,
		DateCreated: curDate,
		UserID:      tAuthor.ID,
	}
	tURLPost = post.Post{
		ID:          uuid.New(),
		Type:        "url",
		Title:       "title",
		Body:        "hhtp://x.com",
		Category:    "url",
		Score:       1,
		Views:       1,
		DateCreated: curDate,
		UserID:      tAuthor.ID,
	}
	tAppPost = toAppPost(tPost, tAuthor, tComments, tCommAuthors, tVotes)
	errFoo   = errors.New("some error")
)

func toAppPostsI[T AppPost](posts []T) []AppPost {
	pss := make([]AppPost, len(posts))
	for i := range posts {
		pss[i] = posts[i]
	}

	return pss
}

type contextData struct {
	route  string
	params map[string]string
}

func (cd contextData) Route() string {
	return cd.route
}

func (cd contextData) Params() map[string]string {
	return cd.params
}

func TestPostsHandler_List(t *testing.T) {
	tests := []struct {
		name              string
		posts             []post.Post
		wantBody          paging.Response[AppPost]
		postsRepoErr      error
		postsCountRepoErr error
		userRepoErr       error
		wantErrMsg        string
		wantStatus        int
		qparams           string
	}{
		{
			name:       "list empty posts",
			posts:      []post.Post{},
			wantBody:   paging.NewResponse([]AppPost{}, 1, 1, 10),
			wantStatus: http.StatusOK,
		},
		{
			name:       "list posts",
			posts:      []post.Post{tPost},
			wantBody:   paging.NewResponse([]AppPost{tAppPost}, 1, 1, 10),
			wantStatus: http.StatusOK,
		},
		{
			name:       "page parse error",
			qparams:    "page=@", // parse page support only ints
			wantErrMsg: "[{\"field\":\"page\",\"error\":\"strconv.Atoi: parsing \\\"@\\\": invalid syntax\"}]",
		},
		{
			name:         "list posts produce an error",
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Errorf("collecting posts: %w", errFoo).Error(),
		},
		{
			name:        "collecting post additional info prduce an error",
			posts:       []post.Post{tPost},
			userRepoErr: errFoo,
			wantErrMsg:  fmt.Errorf("collecting users: %w", errFoo).Error(),
		},
		{
			name:              "counting posts produce an error",
			posts:             []post.Post{tPost},
			postsCountRepoErr: errFoo,
			wantErrMsg:        fmt.Errorf("counting posts: %w", errFoo).Error(),
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()

		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("GetAll", context.Background(), 1, 10).Return(tt.posts, tt.postsRepoErr)
			postUsecase.Mock.On("Count", context.Background()).Return(1, tt.postsCountRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByIDs", context.Background(), mock.Anything).Return([]user.User{tAuthor}, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostIDs", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostIDs", mock.Anything, mock.Anything).Return(tVotes, nil)

			r := httptest.NewRequest(http.MethodGet, "/?"+tt.qparams, nil)
			w := httptest.NewRecorder()

			err := handler.List(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantBody)

			assert.Equal(t, expectedBody, actualBody)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestPostsHandler_ListByCatName(t *testing.T) {
	tests := []struct {
		name              string
		posts             []post.Post
		wantBody          paging.Response[AppPost]
		postsRepoErr      error
		postsCountRepoErr error
		userRepoErr       error
		wantErrMsg        string
		wantStatus        int
		qparams           string
	}{
		{
			name:       "list empty posts",
			posts:      []post.Post{},
			wantBody:   paging.NewResponse([]AppPost{}, 1, 1, 10),
			wantStatus: http.StatusOK,
		},
		{
			name:       "list posts",
			posts:      []post.Post{tPost},
			wantBody:   paging.NewResponse([]AppPost{tAppPost}, 1, 1, 10),
			wantStatus: http.StatusOK,
		},
		{
			name:         "list posts produce an error",
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Errorf("collecting posts by category: %w", errFoo).Error(),
		},
		{
			name:       "page parse error",
			qparams:    "page=@", // parse page support only ints
			wantErrMsg: "[{\"field\":\"page\",\"error\":\"strconv.Atoi: parsing \\\"@\\\": invalid syntax\"}]",
		},
		{
			name:              "counting posts produce an error",
			posts:             []post.Post{tPost},
			postsCountRepoErr: errFoo,
			wantErrMsg:        fmt.Errorf("counting posts: %w", errFoo).Error(),
		},
		{
			name:        "collecting post additional info prduce an error",
			posts:       []post.Post{tPost},
			userRepoErr: errFoo,
			wantErrMsg:  fmt.Errorf("collecting users: %w", errFoo).Error(),
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()

		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("GetByCatName", context.Background(), tPost.Category).Return(tt.posts, tt.postsRepoErr)
			postUsecase.Mock.On("Count", context.Background()).Return(1, tt.postsCountRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByIDs", context.Background(), mock.Anything).Return([]user.User{tAuthor}, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostIDs", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostIDs", mock.Anything, mock.Anything).Return(tVotes, nil)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:category_name",
				params: map[string]string{"category_name": tPost.Category},
			})

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", tt.qparams), nil).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.ListByCatName(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantBody)

			assert.Equal(t, expectedBody, actualBody)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestPostsHandler_ListByUsername(t *testing.T) {
	tests := []struct {
		name              string
		posts             []post.Post
		wantBody          paging.Response[AppPost]
		postsRepoErr      error
		postsCountRepoErr error
		commentsRepoErr   error
		userRepoErr       error
		wantErrMsg        string
		wantStatus        int
		qparams           string
	}{
		{
			name:       "list empty posts",
			posts:      []post.Post{},
			wantBody:   paging.NewResponse([]AppPost{}, 1, 1, 10),
			wantStatus: http.StatusOK,
		},
		{
			name:       "list posts OK",
			posts:      []post.Post{tPost},
			wantBody:   paging.NewResponse([]AppPost{tAppPost}, 1, 1, 10),
			wantStatus: http.StatusOK,
		},
		{
			name:       "page parse error",
			qparams:    "page=@", // parse page support only ints
			wantErrMsg: "[{\"field\":\"page\",\"error\":\"strconv.Atoi: parsing \\\"@\\\": invalid syntax\"}]",
		},
		{
			name:         "list posts produce an error",
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Errorf("collecting posts by username: %w", errFoo).Error(),
		},
		{
			name:        "get user info produce an error",
			posts:       []post.Post{tPost},
			userRepoErr: errFoo,
			wantErrMsg:  fmt.Errorf("getting user: %w", errFoo).Error(),
		},
		{
			name:        "get user info produce user not found error",
			userRepoErr: user.ErrNotFound,
			wantErrMsg:  "not found",
		},
		{
			name:            "collecting post additional info produce an error",
			posts:           []post.Post{tPost},
			commentsRepoErr: errFoo,
			wantErrMsg:      fmt.Errorf("collecting comments: %w", errFoo).Error(),
		},
		{
			name:              "counting posts produce an error",
			posts:             []post.Post{tPost},
			postsCountRepoErr: errFoo,
			wantErrMsg:        fmt.Errorf("counting posts: %w", errFoo).Error(),
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}
		t.Run(tt.name, func(t *testing.T) {
			userUsecase.Mock.On("GetByUsername", context.Background(), tAuthor.Name).Return(tAuthor, tt.userRepoErr)
			postUsecase.Mock.On("GetByUserID", context.Background(), tAuthor.ID).Return(tt.posts, tt.postsRepoErr)
			postUsecase.Mock.On("Count", context.Background()).Return(1, tt.postsCountRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByIDs", context.Background(), mock.Anything).Return([]user.User{tAuthor}, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostIDs", mock.Anything, mock.Anything).Return(tComments, tt.commentsRepoErr)
			postUsecase.Mock.On("GetVotesByPostIDs", mock.Anything, mock.Anything).Return(tVotes, nil)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:user_name",
				params: map[string]string{"user_name": tAuthor.Name},
			})

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/user/%s?%s", tAuthor.Name, tt.qparams), nil).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.ListByUsername(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantBody)

			assert.Equal(t, expectedBody, actualBody)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestPostsHandler_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		post         post.Post
		postID       string
		wantBody     AppPost
		postsRepoErr error
		userRepoErr  error
		wantErrMsg   string
		wantStatus   int
	}{
		{
			name:       "get post",
			post:       tPost,
			postID:     tPost.ID.String(),
			wantBody:   tAppPost,
			wantStatus: http.StatusOK,
		},
		{
			name:       "post id is not in uuid format",
			postID:     "#",
			wantErrMsg: "[{\"field\":\"post_id\",\"error\":\"invalid UUID length: 1\"}]",
		},
		{
			name:         "getting not existing post should produce an error",
			postID:       tPost.ID.String(),
			postsRepoErr: post.ErrNotFound,
			wantErrMsg:   post.ErrNotFound.Error(),
		},
		{
			name:         "error from repo should be thrown",
			postID:       tPost.ID.String(),
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Errorf("getting post(%v): %w", tPost.ID, errFoo).Error(),
		},
		{
			name:        "error from repo should be thrown",
			postID:      tPost.ID.String(),
			userRepoErr: errFoo,
			wantErrMsg:  fmt.Errorf("getting post author: %w", errFoo).Error(),
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()

		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("GetByID", context.Background(), tPost.ID).Return(tt.post, tt.postsRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByID", context.Background(), mock.Anything).Return(tAuthor, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostID", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostID", mock.Anything, mock.Anything).Return(tVotes, nil)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:post_id",
				params: map[string]string{"post_id": tt.postID},
			})
			r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.GetByID(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg, "Should return wrapped error from repo")
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantBody)

			assert.Equal(t, expectedBody, actualBody)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestPostsHandler_AddPost(t *testing.T) {
	np := AppNewPost{
		Title:    "title",
		Type:     "text",
		Text:     "text",
		Category: "category",
	}

	tests := []struct {
		name         string
		np           AppNewPost
		post         post.Post
		wantPost     AppPost
		wantErrMsg   string
		postsRepoErr error
		userRepoErr  error
	}{
		{
			name:     "add new post",
			np:       np,
			post:     tPost,
			wantPost: tAppPost,
		},
		{
			name:       "payload decode error should be thrown",
			wantErrMsg: "unable to decode payload: unable to validate payload: [{\"field\":\"title\",\"error\":\"title is a required field\"},{\"field\":\"type\",\"error\":\"type is a required field\"},{\"field\":\"category\",\"error\":\"category is a required field\"}]",
		},
		{
			name:         "error wrong post type should be thrown",
			np:           np,
			postsRepoErr: post.ErrWrongPostType,
			wantErrMsg:   "new post should be url or text",
		},
		{
			name:         "post add error should be thrown",
			np:           np,
			postsRepoErr: errFoo,
			wantErrMsg:   "creating new post: some error",
		},
		{
			name:        "get extended post info should be thrown",
			np:          np,
			userRepoErr: errFoo,
			wantErrMsg:  "getting post author: some error",
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("Add", mock.Anything, mock.Anything, toCoreNewPost(tt.np), mock.Anything).Return(tt.post, tt.postsRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByID", context.Background(), mock.Anything).Return(tAuthor, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostID", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostID", mock.Anything, mock.Anything).Return(tVotes, nil)

			body, _ := json.Marshal(tt.np)
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			err := handler.AddPost(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantPost)

			assert.Equal(t, expectedBody, actualBody)
		})
	}
}

func TestPostsHandler_DeleteByID(t *testing.T) {
	tests := []struct {
		name         string
		postID       string
		postsRepoErr error
		wantErrMsg   string
		wantStatus   int
	}{
		{
			name:       "delete existing post",
			postID:     tPost.ID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:         "delete error should be thrown",
			postID:       tPost.ID.String(),
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Sprintf("deleting post(%s): some error", tPost.ID),
		},
		{
			name:         "delete not existing post should produce an error",
			postID:       tPost.ID.String(),
			postsRepoErr: post.ErrNotFound,
			wantErrMsg:   post.ErrNotFound.Error(),
		},
		{
			name:         "error on attempt to delete post of different user",
			postID:       tPost.ID.String(),
			postsRepoErr: post.ErrForbidden,
			wantErrMsg:   post.ErrForbidden.Error(),
		},
		{
			name:       "parse postID err shoud be thrown",
			postID:     "x",
			wantErrMsg: "[{\"field\":\"post_id\",\"error\":\"invalid UUID length: 1\"}]",
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(tt.postsRepoErr)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:post_id",
				params: map[string]string{"post_id": tt.postID},
			})
			r := httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.DeleteByID(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg, "Should return wrapped error from repo")
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(web.MessageResponse{Msg: "success"})

			assert.Equal(t, expectedBody, actualBody)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestPostsHandler_AddComment(t *testing.T) {
	newComment := post.NewComment{
		Text: "comment text",
	}

	tests := []struct {
		name         string
		nc           post.NewComment
		post         post.Post
		postID       string
		wantPost     AppPost
		wantErrMsg   string
		postsRepoErr error
		userRepoErr  error
	}{
		{
			name:     "add new comment",
			nc:       newComment,
			post:     tPost,
			postID:   tPost.ID.String(),
			wantPost: tAppPost,
		},
		{
			name:       "parse postID err shoud be thrown",
			nc:         newComment,
			postID:     "x",
			wantErrMsg: "[{\"field\":\"post_id\",\"error\":\"invalid UUID length: 1\"}]",
		},
		{
			name:         "post not exists error",
			nc:           newComment,
			postID:       tPost.ID.String(),
			postsRepoErr: post.ErrNotFound,
			wantErrMsg:   "post not found",
		},
		{
			name:         "create comment error should be thrown",
			nc:           newComment,
			postID:       tPost.ID.String(),
			wantErrMsg:   fmt.Sprintf("creating comment for post(%s): some error", tPost.ID),
			postsRepoErr: errors.New("some error"),
		},
		{
			name:       "payload decode error should be thrown",
			wantErrMsg: "unable to decode payload: unable to validate payload: [{\"field\":\"text\",\"error\":\"text is a required field\"}]",
		},
		{
			name:        "get extended post info should be thrown",
			nc:          newComment,
			postID:      tPost.ID.String(),
			userRepoErr: errFoo,
			wantErrMsg:  "getting post author: some error",
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("AddComment", mock.Anything, mock.Anything, tPost.ID, tt.nc, mock.Anything).Return(tt.post, tt.postsRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByID", context.Background(), mock.Anything).Return(tAuthor, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostID", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostID", mock.Anything, mock.Anything).Return(tVotes, nil)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:post_id",
				params: map[string]string{"post_id": tt.postID},
			})

			body, _ := json.Marshal(tt.nc)
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.AddComment(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantPost)

			assert.Equal(t, expectedBody, actualBody)
		})
	}
}

func TestPostsHandler_DeleteComment(t *testing.T) {
	cuuid := uuid.New()

	tests := []struct {
		name         string
		post         post.Post
		postID       string
		commentID    string
		wantPost     AppPost
		wantStatus   int
		postsRepoErr error
		userRepoErr  error
		wantErrMsg   string
	}{
		{
			name:       "delete existing comment",
			post:       tPost,
			postID:     tPost.ID.String(),
			commentID:  cuuid.String(),
			wantPost:   tAppPost,
			wantStatus: http.StatusOK,
		},
		{
			name:         "delete error should be thrown",
			postID:       tPost.ID.String(),
			commentID:    cuuid.String(),
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Errorf("deleting comment(%s) of post(%s): %w", cuuid.String(), tPost.ID, errFoo).Error(),
		},
		{
			name:       "parse postID err shoud be thrown",
			postID:     "x",
			commentID:  cuuid.String(),
			wantErrMsg: "[{\"field\":\"post_id\",\"error\":\"invalid UUID length: 1\"}]",
		},
		{
			name:       "parse commentID err shoud be thrown",
			commentID:  "x",
			wantErrMsg: "[{\"field\":\"comment_id\",\"error\":\"invalid UUID length: 1\"}]",
		},
		{
			name:         "delete comment of non existing post should produce an error",
			postID:       tPost.ID.String(),
			commentID:    cuuid.String(),
			postsRepoErr: post.ErrNotFound,
			wantErrMsg:   post.ErrNotFound.Error(),
		},
		{
			name:         "delete not existing comment should produce an error",
			postID:       tPost.ID.String(),
			commentID:    cuuid.String(),
			postsRepoErr: post.ErrCommentNotFound,
			wantErrMsg:   post.ErrCommentNotFound.Error(),
		},
		{
			name:         "error on attempt to delete post of different user",
			postID:       tPost.ID.String(),
			commentID:    cuuid.String(),
			postsRepoErr: post.ErrForbidden,
			wantErrMsg:   post.ErrForbidden.Error(),
		},
		{
			name:        "get extended post info should be thrown",
			postID:      tPost.ID.String(),
			commentID:   cuuid.String(),
			userRepoErr: errFoo,
			wantErrMsg:  "getting post author: some error",
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("DeleteComment", mock.Anything, mock.Anything, tPost.ID, cuuid).Return(tt.post, tt.postsRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByID", context.Background(), mock.Anything).Return(tAuthor, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostID", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostID", mock.Anything, mock.Anything).Return(tVotes, nil)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:post_id/:comment_id",
				params: map[string]string{"post_id": tt.postID, "comment_id": tt.commentID},
			})

			r := httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.DeleteComment(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg, "Should return wrapped error from repo")
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantPost)

			assert.Equal(t, expectedBody, actualBody)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestPostsHandler_UpVote(t *testing.T) {
	tests := []struct {
		name         string
		post         post.Post
		postID       string
		wantPost     AppPost
		wantErrMsg   string
		postsRepoErr error
		userRepoErr  error
	}{
		{
			name:     "vote should be count",
			post:     tPost,
			postID:   tPost.ID.String(),
			wantPost: tAppPost,
		},
		{
			name:         "post not exists error",
			postID:       tPost.ID.String(),
			postsRepoErr: post.ErrNotFound,
			wantErrMsg:   "post not found",
		},
		{
			name:         "upvote error should be thrown",
			postID:       tPost.ID.String(),
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Errorf("upvote post(%s): %w", tPost.ID, errFoo).Error(),
		},
		{
			name:        "get extended post info should be thrown",
			postID:      tPost.ID.String(),
			userRepoErr: errFoo,
			wantErrMsg:  "getting post author: some error",
		},
		{
			name:       "parse postID err shoud be thrown",
			postID:     "x",
			wantErrMsg: "[{\"field\":\"post_id\",\"error\":\"invalid UUID length: 1\"}]",
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("AddVote", context.Background(), mock.Anything, tPost.ID, mock.AnythingOfType("int32")).Return(tt.post, tt.postsRepoErr)
			// mock get posts info
			userUsecase.Mock.On("GetByID", context.Background(), mock.Anything).Return(tAuthor, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostID", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostID", mock.Anything, mock.Anything).Return(tVotes, nil)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:post_id",
				params: map[string]string{"post_id": tt.postID},
			})

			r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.UpVote(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantPost)

			assert.Equal(t, expectedBody, actualBody)
		})
	}
}

func TestPostsHandler_DownVote(t *testing.T) {
	tests := []struct {
		name         string
		post         post.Post
		postID       string
		wantPost     AppPost
		wantErrMsg   string
		postsRepoErr error
		userRepoErr  error
	}{
		{
			name:     "vote should be count",
			post:     tPost,
			postID:   tPost.ID.String(),
			wantPost: tAppPost,
		},
		{
			name:         "post not exists error",
			postID:       tPost.ID.String(),
			postsRepoErr: post.ErrNotFound,
			wantErrMsg:   "post not found",
		},
		{
			name:         "vote error in repo should be thrown",
			postID:       tPost.ID.String(),
			postsRepoErr: errFoo,
			wantErrMsg:   fmt.Errorf("downvote post(%s): %w", tPost.ID, errFoo).Error(),
		},
		{
			name:        "get extended post info should be thrown",
			postID:      tPost.ID.String(),
			userRepoErr: errFoo,
			wantErrMsg:  "getting post author: some error",
		},
		{
			name:       "parse postID err shoud be thrown",
			postID:     "x",
			wantErrMsg: "[{\"field\":\"post_id\",\"error\":\"invalid UUID length: 1\"}]",
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			postUsecase.Mock.On("AddVote", mock.Anything, mock.Anything, tPost.ID, mock.AnythingOfType("int32")).Return(tt.post, tt.postsRepoErr).Once()
			// mock get posts info
			userUsecase.Mock.On("GetByID", context.Background(), mock.Anything).Return(tAuthor, tt.userRepoErr)
			postUsecase.Mock.On("GetCommentsByPostID", mock.Anything, mock.Anything).Return(tComments, nil)
			postUsecase.Mock.On("GetVotesByPostID", mock.Anything, mock.Anything).Return(tVotes, nil)

			ctx := httptreemux.AddRouteDataToContext(context.Background(), contextData{
				route:  "/:post_id",
				params: map[string]string{"post_id": tt.postID},
			})

			r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			err := handler.DownVote(context.Background(), w, r)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantPost)

			assert.Equal(t, expectedBody, actualBody)
		})
	}
}

func TestPostsHandler_getPostsInfo(t *testing.T) {
	tVote := post.Vote{
		Vote:   1,
		User:   tPost.UserID,
		PostID: tPost.ID,
	}
	tComment := post.Comment{
		ID:          uuid.New(),
		PostID:      tPost.ID,
		UserID:      uuid.New(),
		DateCreated: time.Now(),
	}

	tests := []struct {
		name         string
		posts        []post.Post
		comments     []post.Comment
		votes        []post.Vote
		wantAppPosts []AppPost

		wantErrMsg         string
		usersGetRepoErr    error
		commGetRepoErr     error
		commUserGetRepoErr error
		votesGetRepoErr    error
	}{
		{
			name:     "collect author, comments and votes",
			posts:    []post.Post{tPost},
			comments: []post.Comment{tComment},
			votes:    []post.Vote{tVote},
			wantAppPosts: toAppPostsI([]AppTextPost{
				{
					ID:               tPost.ID.String(),
					Type:             tPost.Type,
					Title:            tPost.Title,
					Text:             tPost.Body,
					Category:         tPost.Category,
					Score:            tPost.Score,
					Views:            tPost.Views,
					DateCreated:      tPost.DateCreated.Format(time.RFC3339),
					Author:           AppPostAuthor{ID: tAuthor.ID.String()},
					UpvotePercentage: 100,
					Votes: []AppVote{
						{
							Vote: tVote.Vote,
							User: tVote.User.String(),
						},
					},
					Comments: []AppComment{
						{
							ID:     tComment.ID.String(),
							PostID: tComment.PostID.String(),
							Author: AppPostAuthor{
								ID: tComment.UserID.String(),
							},
							DateCreated: tComment.DateCreated.Format(time.RFC3339),
						},
					},
				},
			}),
		},
		{
			name:            "get posts authors produce an error",
			usersGetRepoErr: errFoo,
			wantErrMsg:      fmt.Errorf("collecting users: %w", errFoo).Error(),
			posts:           []post.Post{tPost},
		},
		{
			name:           "get posts comments produce an error",
			commGetRepoErr: errFoo,
			wantErrMsg:     fmt.Errorf("collecting comments: %w", errFoo).Error(),
			posts:          []post.Post{tPost},
		},
		{
			name:               "get comments authors produce an error",
			commUserGetRepoErr: errFoo,
			wantErrMsg:         fmt.Errorf("collecting comment authors: %w", errFoo).Error(),
			posts:              []post.Post{tPost},
		},
		{
			name:            "get votes produce an error",
			votesGetRepoErr: errFoo,
			wantErrMsg:      fmt.Errorf("collecting votes: %w", errFoo).Error(),
			posts:           []post.Post{tPost},
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		pIDs := make([]uuid.UUID, 0, len(tt.posts))
		pAuthorsIDs := make([]uuid.UUID, 0, len(tt.posts))
		pAuthors := make([]user.User, 0, len(tt.posts))
		for _, p := range tt.posts {
			pIDs = append(pIDs, p.ID)
			pAuthors = append(pAuthors, user.User{ID: p.UserID})
			pAuthorsIDs = append(pAuthorsIDs, p.UserID)
		}

		cAuthorsIDs := make([]uuid.UUID, 0, len(tt.comments))
		cAuthors := make([]user.User, 0, len(tt.comments))
		for _, c := range tt.comments {
			cAuthors = append(cAuthors, user.User{ID: c.UserID})
			cAuthorsIDs = append(cAuthorsIDs, c.UserID)
		}

		t.Run(tt.name, func(t *testing.T) {
			userUsecase.Mock.On("GetByIDs", context.Background(), pAuthorsIDs).Return(pAuthors, tt.usersGetRepoErr).Once()
			postUsecase.Mock.On("GetCommentsByPostIDs", context.Background(), pIDs).Return(tt.comments, tt.commGetRepoErr)
			userUsecase.Mock.On("GetByIDs", context.Background(), cAuthorsIDs).Return(cAuthors, tt.commUserGetRepoErr).Once()
			postUsecase.Mock.On("GetVotesByPostIDs", context.Background(), pIDs).Return(tt.votes, tt.votesGetRepoErr)

			posts, err := handler.getPostsInfo(context.Background(), tt.posts)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			assert.Equal(t, tt.wantAppPosts, posts)
		})
	}
}

func TestPostsHandler_getPostInfo(t *testing.T) {
	tVote := post.Vote{
		User:   tPost.UserID,
		PostID: tPost.ID,
	}
	tComment := post.Comment{
		ID:          uuid.New(),
		PostID:      tPost.ID,
		UserID:      uuid.New(),
		DateCreated: time.Now(),
	}

	tests := []struct {
		name        string
		post        post.Post
		comments    []post.Comment
		votes       []post.Vote
		wantAppPost AppPost

		wantErrMsg         string
		usersGetRepoErr    error
		commGetRepoErr     error
		commUserGetRepoErr error
		votesGetRepoErr    error
	}{
		{
			name:     "collect author, comments and votes",
			post:     tURLPost,
			comments: []post.Comment{tComment},
			votes:    []post.Vote{tVote},
			wantAppPost: AppURLPost{
				ID:          tURLPost.ID.String(),
				Type:        tURLPost.Type,
				Title:       tURLPost.Title,
				URL:         tURLPost.Body,
				Category:    tURLPost.Category,
				Score:       tURLPost.Score,
				Views:       tURLPost.Views,
				DateCreated: tURLPost.DateCreated.Format(time.RFC3339),
				Author:      AppPostAuthor{ID: tAuthor.ID.String(), Username: tAuthor.Name},
				Votes: []AppVote{
					{
						Vote: tVote.Vote,
						User: tVote.User.String(),
					},
				},
				Comments: []AppComment{
					{
						ID:     tComment.ID.String(),
						PostID: tComment.PostID.String(),
						Author: AppPostAuthor{
							ID: tComment.UserID.String(),
						},
						DateCreated: tComment.DateCreated.Format(time.RFC3339),
					},
				},
			},
		},
		{
			name:            "get posts authors produce an error",
			usersGetRepoErr: errFoo,
			wantErrMsg:      fmt.Errorf("getting post author: %w", errFoo).Error(),
			post:            tPost,
		},
		{
			name:           "get posts comments produce an error",
			commGetRepoErr: errFoo,
			wantErrMsg:     fmt.Errorf("getting post comments: %w", errFoo).Error(),
			post:           tPost,
		},
		{
			name:               "get comments authors produce an error",
			commUserGetRepoErr: errFoo,
			wantErrMsg:         fmt.Errorf("getting post comments authors: %w", errFoo).Error(),
			post:               tPost,
			comments:           []post.Comment{tComment},
		},
		{
			name:            "get votes produce an error",
			votesGetRepoErr: errFoo,
			wantErrMsg:      fmt.Errorf("getting post votes: %w", errFoo).Error(),
			post:            tPost,
		},
	}

	for _, tt := range tests {
		postUsecase := post.NewUsecaseMock()
		userUsecase := user.NewUsecaseMock()
		handler := &PostsHandler{
			Posts: postUsecase,
			Users: userUsecase,
		}

		cAuthorsIDs := make([]uuid.UUID, 0, len(tt.comments))
		cAuthors := make([]user.User, 0, len(tt.comments))
		for _, c := range tt.comments {
			cAuthors = append(cAuthors, user.User{ID: c.UserID})
			cAuthorsIDs = append(cAuthorsIDs, c.UserID)
		}

		t.Run(tt.name, func(t *testing.T) {
			userUsecase.Mock.On("GetByID", context.Background(), tt.post.UserID).Return(tAuthor, tt.usersGetRepoErr).Once()
			postUsecase.Mock.On("GetCommentsByPostID", context.Background(), tt.post.ID).Return(tt.comments, tt.commGetRepoErr)
			userUsecase.Mock.On("GetByIDs", context.Background(), cAuthorsIDs).Return(cAuthors, tt.commUserGetRepoErr).Once()
			postUsecase.Mock.On("GetVotesByPostID", context.Background(), tt.post.ID).Return(tt.votes, tt.votesGetRepoErr)

			posts, err := handler.getPostInfo(context.Background(), tt.post)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			assert.Equal(t, tt.wantAppPost, posts)
		})
	}
}
