package post

import (
	"context"
	"time"

	"github.com/rocketb/asperitas/internal/web/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type UsecaseMock struct {
	mock.Mock
}

func NewUsecaseMock() *UsecaseMock {
	return &UsecaseMock{}
}

func (r *UsecaseMock) GetAll(ctx context.Context, pageNum int, rowsPerPage int) ([]Post, error) {
	args := r.Called(ctx, pageNum, rowsPerPage)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Post), args.Error(1)
}

func (r *UsecaseMock) GetByID(ctx context.Context, postID uuid.UUID) (Post, error) {
	args := r.Called(ctx, postID)
	if args.Get(1) != nil {
		return Post{}, args.Error(1)
	}

	return args.Get(0).(Post), args.Error(1)
}

func (r *UsecaseMock) GetByCatName(ctx context.Context, catName string) ([]Post, error) {
	args := r.Called(ctx, catName)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Post), args.Error(1)
}

func (r *UsecaseMock) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error) {
	args := r.Called(ctx, userID)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Post), args.Error(1)
}

func (r *UsecaseMock) Add(ctx context.Context, claims auth.Claims, np NewPost, now time.Time) (Post, error) {
	args := r.Called(ctx, claims, np, now)
	if args.Get(1) != nil {
		return Post{}, args.Error(1)
	}

	return args.Get(0).(Post), args.Error(1)
}

func (r *UsecaseMock) Count(ctx context.Context) (int, error) {
	args := r.Called(ctx)
	if args.Get(1) != nil {
		return 0, args.Error(1)
	}

	return args.Get(0).(int), args.Error(1)
}

func (r *UsecaseMock) Delete(ctx context.Context, claims auth.Claims, postID uuid.UUID) error {
	args := r.Called(ctx, claims, postID)
	return args.Error(0)
}

func (r *UsecaseMock) AddVote(ctx context.Context, claims auth.Claims, postID uuid.UUID, vote int32) (Post, error) {
	args := r.Called(ctx, claims, postID, vote)
	if args.Get(1) != nil {
		return Post{}, args.Error(1)
	}

	return args.Get(0).(Post), args.Error(1)
}

func (r *UsecaseMock) GetVotesByPostID(ctx context.Context, postID uuid.UUID) ([]Vote, error) {
	args := r.Called(ctx, postID)
	if args.Get(1) != nil {
		return []Vote{}, args.Error(1)
	}

	return args.Get(0).([]Vote), args.Error(1)
}

func (r *UsecaseMock) GetVotesByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Vote, error) {
	args := r.Called(ctx, postIDs)
	if args.Get(1) != nil {
		return []Vote{}, args.Error(1)
	}

	return args.Get(0).([]Vote), args.Error(1)
}

func (r *UsecaseMock) AddComment(ctx context.Context, claims auth.Claims, postID uuid.UUID, nc NewComment, now time.Time) (Post, error) {
	args := r.Called(ctx, claims, postID, nc, now)
	if args.Get(1) != nil {
		return Post{}, args.Error(1)
	}

	return args.Get(0).(Post), args.Error(1)
}

func (r *UsecaseMock) GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]Comment, error) {
	args := r.Called(ctx, postID)
	if args.Get(1) != nil {
		return []Comment{}, args.Error(1)
	}

	return args.Get(0).([]Comment), args.Error(1)
}

func (r *UsecaseMock) GetCommentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Comment, error) {
	args := r.Called(ctx, postIDs)
	if args.Get(1) != nil {
		return []Comment{}, args.Error(1)
	}

	return args.Get(0).([]Comment), args.Error(1)
}

func (r *UsecaseMock) DeleteComment(ctx context.Context, claims auth.Claims, postID, commentID uuid.UUID) (Post, error) {
	args := r.Called(ctx, claims, postID, commentID)
	if args.Get(1) != nil {
		return Post{}, args.Error(1)
	}

	return args.Get(0).(Post), args.Error(1)
}
