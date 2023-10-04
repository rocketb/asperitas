package post

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/internal/web/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	puuid = uuid.New()
	tUser = user.User{
		ID:   puuid,
		Name: "uname",
	}
	tPost = Post{
		ID:       puuid,
		Type:     "text",
		Title:    "title",
		Category: "category",
		UserID:   tUser.ID,
	}
	tVote    = Vote{}
	tComment = Comment{
		ID:          uuid.New(),
		DateCreated: time.Now(),
		Body:        "comment",
		UserID:      tUser.ID,
	}
	tPostInfo = Post{
		ID:       tPost.ID,
		Type:     tPost.Type,
		Title:    tPost.Title,
		Body:     tPost.Body,
		Category: tPost.Category,
	}
	errFoo = errors.New("some error")
)

func TestGetAll(t *testing.T) {
	tests := []struct {
		name        string
		posts       []Post
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name:    "list all",
			posts:   []Post{tPost},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			posts:       []Post{},
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetAll", context.Background(), 1, 1).Return(tt.posts, tt.postRepoErr)

			posts, err := uc.GetAll(context.Background(), 1, 1)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.posts, posts)
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name        string
		total       int
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name:    "count posts ok",
			wantErr: assert.NoError,
			total:   1,
		},
		{
			name:        "error on count posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("Count", context.Background()).Return(tt.total, tt.postRepoErr)

			total, err := uc.Count(context.Background())
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.total, total)
		})
	}
}

func TestGetByCatname(t *testing.T) {
	tests := []struct {
		name        string
		posts       []Post
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name:    "list all",
			posts:   []Post{tPost},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetByCatName", context.Background(), tPost.Category).Return(tt.posts, tt.postRepoErr)

			posts, err := uc.GetByCatName(context.Background(), tPost.Category)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.posts, posts)
		})
	}
}

func TestGetByUserID(t *testing.T) {
	tests := []struct {
		name        string
		posts       []Post
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
		userRepoErr error
		caseErr     error
	}{
		{
			name:    "list all",
			posts:   []Post{tPost},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			caseErr:     errFoo,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetByUserID", context.Background(), tUser.ID).Return(tt.posts, tt.postRepoErr)
			posts, err := uc.GetByUserID(context.Background(), tUser.ID)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.caseErr)
			}

			assert.Equal(t, tt.posts, posts)
		})
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name        string
		post        Post
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name:    "get by id",
			post:    tPost,
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tt.post, tt.postRepoErr).Once()

			post, err := uc.GetByID(context.Background(), tPost.ID)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.post, post)
		})
	}
}

func TestAddPost(t *testing.T) {
	idGen := func() uuid.UUID {
		return tPost.ID
	}

	curTime := time.Now()
	type args struct {
		claims auth.Claims
		np     NewPost
		now    time.Time
	}
	tests := []struct {
		name        string
		args        args
		wantPost    Post
		wantErr     assert.ErrorAssertionFunc
		caseErr     error
		postRepoErr error
		voteAddErr  error
	}{
		{
			name: "url post",
			args: args{
				np: NewPost{
					Type: "url",
					URL:  "url",
				},
				now: curTime,
				claims: auth.Claims{
					User: auth.User{
						ID: tUser.ID,
					},
				},
			},
			wantPost: Post{
				ID:          tPost.ID,
				Type:        "url",
				Body:        "url",
				DateCreated: curTime,
				UserID:      tUser.ID,
				Score:       1,
			},
			wantErr: assert.NoError,
		},
		{
			name: "text post",
			args: args{
				np: NewPost{
					Type: "text",
					Text: "text",
				},
				claims: auth.Claims{
					User: auth.User{
						ID: tUser.ID,
					},
				},
			},
			wantPost: Post{
				ID:          tPost.ID,
				Type:        "text",
				Body:        "text",
				DateCreated: curTime,
				UserID:      tUser.ID,
				Score:       1,
			},
			wantErr: assert.NoError,
		},
		{
			name: "error wrong post type",
			args: args{
				np: NewPost{
					Type: "not_exist",
				},
			},
			wantErr: assert.Error,
			caseErr: ErrWrongPostType,
		},
		{
			name: "error adding post",
			args: args{
				np: NewPost{
					Type: "text",
				},
				claims: auth.Claims{
					User: auth.User{
						ID: tUser.ID,
					},
				},
			},
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			caseErr:     errFoo,
			wantPost:    Post{},
		},
		{
			name: "add vote error",
			args: args{
				np: NewPost{
					Type: "text",
				},
				claims: auth.Claims{
					User: auth.User{
						ID: tUser.ID,
					},
				},
			},
			wantErr:    assert.Error,
			voteAddErr: errFoo,
			wantPost:   Post{},
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := Core{
			idGen:     idGen,
			PostsRepo: postRepo,
		}

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("Add", context.Background(), mock.Anything).Return(tt.postRepoErr)
			postRepo.Mock.On("AddVote", context.Background(), mock.Anything, mock.Anything).Return(tt.voteAddErr)

			post, err := uc.Add(context.Background(), tt.args.claims, tt.args.np, curTime)
			if tt.wantErr(t, err) {
				if tt.caseErr != nil {
					assert.Equal(t, tt.caseErr, err)
				}
			}
			assert.Equal(t, tt.wantPost, post)
		})
	}
}

func TestDeletePost(t *testing.T) {
	tests := []struct {
		name        string
		post        Post
		claims      auth.Claims
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
		caseErr     error
	}{
		{
			name: "delete post",
			post: tPost,
			claims: auth.Claims{
				User: auth.User{
					ID: tPost.UserID,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "try to delete post of some body else",
			post: tPost,
			claims: auth.Claims{
				User: auth.User{
					ID: uuid.New(),
				},
			},
			wantErr: assert.Error,
			caseErr: ErrForbidden,
		},
		{
			name:        "error on post delete",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			caseErr:     errFoo,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tt.post, tt.postRepoErr).Once()
			postRepo.Mock.On("Delete", context.Background(), tPost.ID).Return(nil).Once()

			err := uc.Delete(context.Background(), tt.claims, tPost.ID)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.caseErr)
			}
		})
	}
}

func TestAddVote(t *testing.T) {
	claims := auth.Claims{
		User: auth.User{ID: tPost.UserID},
	}
	tests := []struct {
		name            string
		wantErr         assert.ErrorAssertionFunc
		getPostErr      error
		addVoteErr      error
		caseErr         error
		checkVoteErr    error
		updateVoteErr   error
		getPostAfterErr error
	}{
		{
			name:         "vote add",
			wantErr:      assert.NoError,
			checkVoteErr: ErrNotFound,
		},
		{
			name:         "vote update",
			wantErr:      assert.NoError,
			checkVoteErr: nil,
		},
		{
			name:       "error on get post",
			wantErr:    assert.Error,
			getPostErr: errFoo,
			caseErr:    errFoo,
		},
		{
			name:         "error on check vote",
			wantErr:      assert.Error,
			checkVoteErr: errFoo,
			caseErr:      errFoo,
		},
		{
			name:         "error on add vote",
			wantErr:      assert.Error,
			checkVoteErr: ErrNotFound,
			addVoteErr:   errFoo,
			caseErr:      errFoo,
		},
		{
			name:          "error on update vote",
			wantErr:       assert.Error,
			updateVoteErr: errFoo,
			caseErr:       errFoo,
		},
		{
			name:            "error on get post after add vote",
			wantErr:         assert.Error,
			getPostAfterErr: errFoo,
			caseErr:         errFoo,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tPost, tt.getPostErr).Once()
			postRepo.Mock.On("CheckVote", context.Background(), tPost.ID, tUser.ID).Return(tt.checkVoteErr)
			postRepo.Mock.On("AddVote", context.Background(), tPost.ID, Vote{Vote: 1, User: tUser.ID}).Return(tt.addVoteErr)
			postRepo.Mock.On("UpdateVote", context.Background(), tPost.ID, mock.Anything).Return(tt.updateVoteErr)
			postRepo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tPost, tt.getPostAfterErr)

			_, err := uc.AddVote(context.Background(), claims, tPost.ID, 1)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.caseErr)
			}
		})
	}
}

func TestGetVotesByPostID(t *testing.T) {
	tests := []struct {
		name        string
		votes       []Vote
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name: "get by id",
			votes: []Vote{
				{
					PostID: tPost.ID,
					Vote:   1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			votes:       []Vote{},
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetVotesByPostID", context.Background(), tPost.ID).Return(tt.votes, tt.postRepoErr).Once()

			votes, err := uc.GetVotesByPostID(context.Background(), tPost.ID)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.votes, votes)
		})
	}
}

func TestGetVotesByPostIDs(t *testing.T) {
	tests := []struct {
		name        string
		votes       []Vote
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name: "get by id",
			votes: []Vote{
				{
					PostID: tPost.ID,
					Vote:   1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			votes:       []Vote{},
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetVotesByPostIDs", context.Background(), []uuid.UUID{tPost.ID}).Return(tt.votes, tt.postRepoErr).Once()

			votes, err := uc.GetVotesByPostIDs(context.Background(), []uuid.UUID{tPost.ID})
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.votes, votes)
		})
	}
}

func TestAddComment(t *testing.T) {
	curTume := time.Now()
	claims := auth.Claims{
		User: auth.User{ID: tPost.UserID},
	}
	newComment := NewComment{
		Text: "text",
	}

	tests := []struct {
		name               string
		wantPost           Post
		wantErr            assert.ErrorAssertionFunc
		addCommentErr      error
		getPostErr         error
		getPostAfterAddErr error
		caseErr            error
	}{
		{
			name:    "comment add",
			wantErr: assert.NoError,
		},
		{
			name:       "error on get post",
			wantErr:    assert.Error,
			getPostErr: errFoo,
			caseErr:    errFoo,
		},
		{
			name:               "error on get post after add",
			wantErr:            assert.Error,
			getPostAfterAddErr: errFoo,
			caseErr:            errFoo,
		},
		{
			name:          "error on add comment",
			wantErr:       assert.Error,
			addCommentErr: errFoo,
			caseErr:       errFoo,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tPost, tt.getPostErr).Once()
			postRepo.Mock.On("AddComment", context.Background(), mock.Anything).Return(tt.addCommentErr)
			postRepo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tPost, tt.getPostAfterAddErr)

			_, err := uc.AddComment(context.Background(), claims, tPost.ID, newComment, curTume)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.caseErr)
			}
		})
	}
}

func TestGetCommentsByPostID(t *testing.T) {
	tests := []struct {
		name        string
		comments    []Comment
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name: "get by id",
			comments: []Comment{
				{
					PostID: tPost.ID,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			comments:    []Comment{},
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetCommentsByPostID", context.Background(), tPost.ID).Return(tt.comments, tt.postRepoErr).Once()

			comments, err := uc.GetCommentsByPostID(context.Background(), tPost.ID)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.comments, comments)
		})
	}
}

func TestGetCommentsByPostIDs(t *testing.T) {
	tests := []struct {
		name        string
		comments    []Comment
		wantErr     assert.ErrorAssertionFunc
		postRepoErr error
	}{
		{
			name: "get by id",
			comments: []Comment{
				{
					PostID: tPost.ID,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting posts",
			wantErr:     assert.Error,
			postRepoErr: errFoo,
			comments:    []Comment{},
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetCommentsByPostIDs", context.Background(), []uuid.UUID{tPost.ID}).Return(tt.comments, tt.postRepoErr).Once()

			comments, err := uc.GetCommentsByPostIDs(context.Background(), []uuid.UUID{tPost.ID})
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.postRepoErr)
			}

			assert.Equal(t, tt.comments, comments)
		})
	}
}
func TestDeleteComment(t *testing.T) {
	type args struct {
		claims    auth.Claims
		postID    uuid.UUID
		commentID uuid.UUID
	}
	tests := []struct {
		name             string
		args             args
		comment          Comment
		wantErr          assert.ErrorAssertionFunc
		caseErr          error
		getCommentErr    error
		deleteCommentErr error
		getPostErr       error
		getPostAfterErr  error
	}{
		{
			name:    "comment delete",
			wantErr: assert.NoError,
			args: args{
				claims: auth.Claims{
					User: auth.User{
						ID: tUser.ID,
					},
				},
				postID:    tPost.ID,
				commentID: tComment.ID,
			},
			comment: tComment,
		},
		{
			name:          "error on get comment",
			wantErr:       assert.Error,
			caseErr:       errFoo,
			getCommentErr: errFoo,
		},
		{
			name: "error on delete comment",
			args: args{
				claims: auth.Claims{
					User: auth.User{
						ID: tPost.UserID,
					},
				},
			},
			wantErr:          assert.Error,
			caseErr:          errFoo,
			deleteCommentErr: errFoo,
			comment:          tComment,
		},
		{
			name: "delete comment of other user",
			args: args{
				claims: auth.Claims{
					User: auth.User{
						ID: uuid.New(),
					},
				},
			},
			wantErr: assert.Error,
			caseErr: ErrForbidden,
		},
		{
			name: "error on get post",
			args: args{
				claims: auth.Claims{
					User: auth.User{
						ID: tPost.UserID,
					},
				},
			},
			wantErr:    assert.Error,
			caseErr:    errFoo,
			getPostErr: errFoo,
			comment:    tComment,
		},
		{
			name: "error on add get post after delete",
			args: args{
				claims: auth.Claims{
					User: auth.User{
						ID: tPost.UserID,
					},
				},
			},
			wantErr:         assert.Error,
			caseErr:         errFoo,
			getPostAfterErr: errFoo,
			comment:         tComment,
		},
	}

	for _, tt := range tests {
		postRepo := NewRepoMock()
		uc := NewCore(postRepo)

		t.Run(tt.name, func(t *testing.T) {
			postRepo.Mock.On("GetByID", context.Background(), tt.args.postID).Return(tPost, tt.getPostErr).Once()
			postRepo.Mock.On("GetCommentByID", context.Background(), tt.args.commentID).Return(tt.comment, tt.getCommentErr)
			postRepo.Mock.On("DeleteComment", context.Background(), tt.args.commentID).Return(tt.deleteCommentErr)
			postRepo.Mock.On("GetByID", context.Background(), tt.args.postID).Return(tPost, tt.getPostAfterErr)

			_, err := uc.DeleteComment(context.Background(), tt.args.claims, tt.args.postID, tt.args.commentID)
			if tt.wantErr(t, err) {
				assert.Equal(t, tt.caseErr, err)
			}
		})
	}
}
