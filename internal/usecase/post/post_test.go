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
	errFoo  = errors.New("some error")
	curTime = time.Now()
)

func TestGetAll(t *testing.T) {
	tests := []struct {
		name  string
		posts []Post
		err   error
	}{
		{
			name:  "list all",
			posts: []Post{tPost},
		},
		{
			name:  "error on getting posts",
			err:   errFoo,
			posts: []Post{},
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetAll", context.Background(), 1, 1).Return(tt.posts, tt.err)

			posts, err := uc.GetAll(context.Background(), 1, 1)
			assert.Equal(t, err, tt.err)
			assert.Equal(t, tt.posts, posts)
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name  string
		total int
		err   error
	}{
		{
			name:  "count posts ok",
			total: 1,
		},
		{
			name: "error on count posts",
			err:  errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("Count", context.Background()).Return(tt.total, tt.err)

			total, err := uc.Count(context.Background())
			assert.Equal(t, err, tt.err)
			assert.Equal(t, tt.total, total)
		})
	}
}

func TestGetByCatname(t *testing.T) {
	tests := []struct {
		name  string
		posts []Post
		err   error
	}{
		{
			name:  "list all",
			posts: []Post{tPost},
		},
		{
			name: "error on getting posts",
			err:  errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByCatName", context.Background(), tPost.Category).Return(tt.posts, tt.err)

			posts, err := uc.GetByCatName(context.Background(), tPost.Category)
			assert.Equal(t, err, tt.err)
			assert.Equal(t, tt.posts, posts)
		})
	}
}

func TestGetByUserID(t *testing.T) {
	tests := []struct {
		name  string
		posts []Post
		err   error
	}{
		{
			name:  "list all ok",
			posts: []Post{tPost},
		},
		{
			name: "error on getting posts",
			err:  errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByUserID", context.Background(), tUser.ID).Return(tt.posts, tt.err)

			posts, err := uc.GetByUserID(context.Background(), tUser.ID)
			assert.Equal(t, err, tt.err)
			assert.Equal(t, tt.posts, posts)
		})
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name string
		post Post
		err  error
	}{
		{
			name: "get by id ok",
			post: tPost,
		},
		{
			name: "error on getting posts",
			err:  errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tt.post, tt.err)

			post, err := uc.GetByID(context.Background(), tPost.ID)
			assert.Equal(t, err, tt.err)
			assert.Equal(t, tt.post, post)
		})
	}
}

func TestAddPost(t *testing.T) {
	type args struct {
		claims auth.Claims
		np     NewPost
		now    time.Time
	}

	tests := []struct {
		name     string
		args     args
		wantPost Post
		caseErr  error
		repoErr  error
		voteErr  error
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
				Type:        "url",
				Body:        "url",
				DateCreated: curTime,
				UserID:      tUser.ID,
				Score:       1,
			},
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
				Type:        "text",
				Body:        "text",
				DateCreated: curTime,
				UserID:      tUser.ID,
				Score:       1,
			},
		},
		{
			name: "error wrong post type",
			args: args{
				np: NewPost{
					Type: "not_exist",
				},
			},
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
			repoErr:  errFoo,
			caseErr:  errFoo,
			wantPost: Post{},
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
			voteErr:  errFoo,
			caseErr:  errFoo,
			wantPost: Post{},
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := Core{
			idGen:     func() uuid.UUID { return tt.wantPost.ID },
			PostsRepo: repo,
		}

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("Add", context.Background(), mock.Anything).Return(tt.repoErr)
			repo.Mock.On("AddVote", context.Background(), mock.Anything, mock.Anything).Return(tt.voteErr)

			post, err := uc.Add(context.Background(), tt.args.claims, tt.args.np, curTime)
			assert.Equal(t, tt.caseErr, err)
			assert.Equal(t, tt.wantPost, post)
		})
	}
}

func TestDeletePost(t *testing.T) {
	tests := []struct {
		name    string
		post    Post
		claims  auth.Claims
		repoErr error
		caseErr error
	}{
		{
			name: "delete post ok",
			post: tPost,
			claims: auth.Claims{
				User: auth.User{
					ID: tPost.UserID,
				},
			},
		},
		{
			name: "try to delete post of other user",
			post: tPost,
			claims: auth.Claims{
				User: auth.User{
					ID: uuid.New(),
				},
			},
			caseErr: ErrForbidden,
		},
		{
			name:    "error on post delete",
			repoErr: errFoo,
			caseErr: errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tt.post, tt.repoErr)
			repo.Mock.On("Delete", context.Background(), tPost.ID).Return(nil)

			err := uc.Delete(context.Background(), tt.claims, tPost.ID)
			assert.Equal(t, err, tt.caseErr)
		})
	}
}

func TestAddVote(t *testing.T) {
	claims := auth.Claims{
		User: auth.User{ID: tPost.UserID},
	}
	tests := []struct {
		name            string
		caseErr         error
		getPostErr      error
		addVoteErr      error
		checkVoteErr    error
		updateVoteErr   error
		getPostAfterErr error
	}{
		{
			name:         "vote add",
			checkVoteErr: ErrNotFound,
		},
		{
			name:         "vote update",
			checkVoteErr: nil,
		},
		{
			name:       "error on get post",
			getPostErr: errFoo,
			caseErr:    errFoo,
		},
		{
			name:         "error on check vote",
			checkVoteErr: errFoo,
			caseErr:      errFoo,
		},
		{
			name:         "error on add vote",
			checkVoteErr: ErrNotFound,
			addVoteErr:   errFoo,
			caseErr:      errFoo,
		},
		{
			name:          "error on update vote",
			updateVoteErr: errFoo,
			caseErr:       errFoo,
		},
		{
			name:            "error on get post after add vote",
			getPostAfterErr: errFoo,
			caseErr:         errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tPost, tt.getPostErr).Once()
			repo.Mock.On("CheckVote", context.Background(), tPost.ID, tUser.ID).Return(tt.checkVoteErr)
			repo.Mock.On("AddVote", context.Background(), tPost.ID, Vote{Vote: 1, User: tUser.ID}).Return(tt.addVoteErr)
			repo.Mock.On("UpdateVote", context.Background(), tPost.ID, mock.Anything).Return(tt.updateVoteErr)
			repo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tPost, tt.getPostAfterErr)

			_, err := uc.AddVote(context.Background(), claims, tPost.ID, 1)
			assert.Equal(t, err, tt.caseErr)
		})
	}
}

func TestGetVotesByPostID(t *testing.T) {
	tests := []struct {
		name  string
		votes []Vote
		err   error
	}{
		{
			name: "get by id",
			votes: []Vote{
				{
					PostID: tPost.ID,
					Vote:   1,
				},
			},
		},
		{
			name:  "error on getting posts",
			err:   errFoo,
			votes: []Vote{},
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetVotesByPostID", context.Background(), tPost.ID).Return(tt.votes, tt.err).Once()

			votes, err := uc.GetVotesByPostID(context.Background(), tPost.ID)
			assert.Equal(t, err, tt.err)
			assert.Equal(t, tt.votes, votes)
		})
	}
}

func TestGetVotesByPostIDs(t *testing.T) {
	tests := []struct {
		name  string
		votes []Vote
		err   error
	}{
		{
			name: "get by id",
			votes: []Vote{
				{
					PostID: tPost.ID,
					Vote:   1,
				},
			},
		},
		{
			name:  "error on getting posts",
			err:   errFoo,
			votes: []Vote{},
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetVotesByPostIDs", context.Background(), []uuid.UUID{tPost.ID}).Return(tt.votes, tt.err).Once()

			votes, err := uc.GetVotesByPostIDs(context.Background(), []uuid.UUID{tPost.ID})
			assert.Equal(t, err, tt.err)
			assert.Equal(t, tt.votes, votes)
		})
	}
}

func TestAddComment(t *testing.T) {
	claims := auth.Claims{
		User: auth.User{ID: tPost.UserID},
	}
	newComment := NewComment{
		Text: "text",
	}

	tests := []struct {
		name               string
		wantPost           Post
		caseErr            error
		addCommentErr      error
		getPostErr         error
		getPostAfterAddErr error
	}{
		{
			name:     "comment add",
			wantPost: tPost,
		},
		{
			name:       "error on get post",
			getPostErr: errFoo,
			caseErr:    errFoo,
		},
		{
			name:               "error on get post after add",
			getPostAfterAddErr: errFoo,
			caseErr:            errFoo,
		},
		{
			name:          "error on add comment",
			addCommentErr: errFoo,
			caseErr:       errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tt.wantPost, tt.getPostErr).Once()
			repo.Mock.On("AddComment", context.Background(), mock.Anything).Return(tt.addCommentErr)
			repo.Mock.On("GetByID", context.Background(), tPost.ID).Return(tPost, tt.getPostAfterAddErr)

			_, err := uc.AddComment(context.Background(), claims, tPost.ID, newComment, curTime)
			assert.Equal(t, err, tt.caseErr)
		})
	}
}

func TestGetCommentsByPostID(t *testing.T) {
	tests := []struct {
		name     string
		comments []Comment
		wantErr  assert.ErrorAssertionFunc
		repoErr  error
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
			name:     "error on getting posts",
			wantErr:  assert.Error,
			repoErr:  errFoo,
			comments: []Comment{},
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetCommentsByPostID", context.Background(), tPost.ID).Return(tt.comments, tt.repoErr).Once()

			comments, err := uc.GetCommentsByPostID(context.Background(), tPost.ID)
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.repoErr)
			}

			assert.Equal(t, tt.comments, comments)
		})
	}
}

func TestGetCommentsByPostIDs(t *testing.T) {
	tests := []struct {
		name     string
		comments []Comment
		wantErr  assert.ErrorAssertionFunc
		repoErr  error
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
			name:     "error on getting posts",
			wantErr:  assert.Error,
			repoErr:  errFoo,
			comments: []Comment{},
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetCommentsByPostIDs", context.Background(), []uuid.UUID{tPost.ID}).Return(tt.comments, tt.repoErr).Once()

			comments, err := uc.GetCommentsByPostIDs(context.Background(), []uuid.UUID{tPost.ID})
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.repoErr)
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
		caseErr          error
		getCommentErr    error
		deleteCommentErr error
		getPostErr       error
		getPostAfterErr  error
	}{
		{
			name: "comment delete",
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
			comment:          tComment,
			caseErr:          errFoo,
			deleteCommentErr: errFoo,
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
			comment:    tComment,
			caseErr:    errFoo,
			getPostErr: errFoo,
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
			comment:         tComment,
			caseErr:         errFoo,
			getPostAfterErr: errFoo,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByID", context.Background(), tt.args.postID).Return(tPost, tt.getPostErr).Once()
			repo.Mock.On("GetCommentByID", context.Background(), tt.args.commentID).Return(tt.comment, tt.getCommentErr)
			repo.Mock.On("DeleteComment", context.Background(), tt.args.commentID).Return(tt.deleteCommentErr)
			repo.Mock.On("GetByID", context.Background(), tt.args.postID).Return(tPost, tt.getPostAfterErr)

			_, err := uc.DeleteComment(context.Background(), tt.args.claims, tt.args.postID, tt.args.commentID)
			assert.Equal(t, tt.caseErr, err)
		})
	}
}
