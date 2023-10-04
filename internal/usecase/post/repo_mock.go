package post

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type RepoMock struct {
	mock.Mock
}

func NewRepoMock() *RepoMock {
	return &RepoMock{}
}

func (r *RepoMock) Add(ctx context.Context, newPost Post) error {
	args := r.Called(ctx, newPost)
	return args.Error(0)
}

func (r *RepoMock) Count(ctx context.Context) (int, error) {
	args := r.Called(ctx)
	if args.Get(1) != nil {
		return 0, args.Error(1)
	}

	return args.Get(0).(int), args.Error(1)
}

func (r *RepoMock) GetAll(ctx context.Context, pageNum, rowsPerPage int) ([]Post, error) {
	args := r.Called(ctx, pageNum, rowsPerPage)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Post), args.Error(1)
}

func (r *RepoMock) GetByID(ctx context.Context, postID uuid.UUID) (Post, error) {
	args := r.Called(ctx, postID)
	if args.Get(1) != nil {
		return Post{}, args.Error(1)
	}

	return args.Get(0).(Post), args.Error(1)
}

func (r *RepoMock) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error) {
	args := r.Called(ctx, userID)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Post), args.Error(1)
}

func (r *RepoMock) GetByCatName(ctx context.Context, catName string) ([]Post, error) {
	args := r.Called(ctx, catName)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Post), args.Error(1)
}

func (r *RepoMock) Delete(ctx context.Context, postID uuid.UUID) error {
	args := r.Called(ctx, postID)
	return args.Error(0)
}

func (r *RepoMock) AddComment(ctx context.Context, newComment Comment) error {
	args := r.Called(ctx, newComment)
	return args.Error(0)
}

func (r *RepoMock) GetCommentByID(ctx context.Context, commentID uuid.UUID) (Comment, error) {
	args := r.Called(ctx, commentID)
	if args.Get(1) != nil {
		return Comment{}, args.Error(1)
	}

	return args.Get(0).(Comment), args.Error(1)
}

func (r *RepoMock) GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]Comment, error) {
	args := r.Called(ctx, postID)
	if args.Get(1) != nil {
		return []Comment{}, args.Error(1)
	}

	return args.Get(0).([]Comment), args.Error(1)
}

func (r *RepoMock) GetCommentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Comment, error) {
	args := r.Called(ctx, postIDs)
	if args.Get(1) != nil {
		return []Comment{}, args.Error(1)
	}

	return args.Get(0).([]Comment), args.Error(1)
}

func (r *RepoMock) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	args := r.Called(ctx, commentID)
	return args.Error(0)
}

func (r *RepoMock) AddVote(ctx context.Context, postID uuid.UUID, vote Vote) error {
	args := r.Called(ctx, postID, vote)
	return args.Error(0)
}

func (r *RepoMock) GetVotesByPostID(ctx context.Context, postID uuid.UUID) ([]Vote, error) {
	args := r.Called(ctx, postID)
	if args.Get(1) != nil {
		return []Vote{}, args.Error(1)
	}

	return args.Get(0).([]Vote), args.Error(1)
}

func (r *RepoMock) GetVotesByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Vote, error) {
	args := r.Called(ctx, postIDs)
	if args.Get(1) != nil {
		return []Vote{}, args.Error(1)
	}

	return args.Get(0).([]Vote), args.Error(1)
}

func (r *RepoMock) UpdateVote(ctx context.Context, postID uuid.UUID, vote Vote) error {
	args := r.Called(ctx, postID, vote)
	return args.Error(0)
}

func (r *RepoMock) CheckVote(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	args := r.Called(ctx, postID, userID)
	return args.Error(0)
}
