package post

import (
	"context"
	"time"

	"github.com/rocketb/asperitas/internal/web/auth"

	"github.com/google/uuid"
)

// Post represents post.
type Post struct {
	ID          uuid.UUID
	Type        string
	Title       string
	Body        string
	Category    string
	Score       int32
	Views       int
	DateCreated time.Time
	UserID      uuid.UUID
}

// NewPost is what we require from user to add a Post.
type NewPost struct {
	Title    string
	Type     string
	Text     string
	URL      string
	Category string
}

// Vote represents info about post votes.
type Vote struct {
	Vote   int32
	User   uuid.UUID
	PostID uuid.UUID
}

// Comment represents info about post comments.
type Comment struct {
	ID          uuid.UUID
	PostID      uuid.UUID
	DateCreated time.Time
	UserID      uuid.UUID
	Body        string
}

// NewComment is what we require from user to add a Comment.
type NewComment struct {
	Text string
}

// Repo represents post storage interface.
type Repo interface {
	Add(ctx context.Context, newPost Post) error
	Count(ctx context.Context) (int, error)
	GetAll(ctx context.Context, pageNum int, rowsPerPage int) ([]Post, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error)
	GetByCatName(ctx context.Context, catName string) ([]Post, error)
	GetByID(ctx context.Context, postID uuid.UUID) (Post, error)
	Delete(ctx context.Context, postID uuid.UUID) error
	AddComment(ctx context.Context, newComment Comment) error
	GetCommentByID(ctx context.Context, commentID uuid.UUID) (Comment, error)
	GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]Comment, error)
	GetCommentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Comment, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error
	AddVote(ctx context.Context, postID uuid.UUID, vote Vote) error
	GetVotesByPostID(ctx context.Context, postID uuid.UUID) ([]Vote, error)
	GetVotesByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Vote, error)
	UpdateVote(cxt context.Context, postID uuid.UUID, vote Vote) error
	CheckVote(cxt context.Context, postID uuid.UUID, userID uuid.UUID) error
}

// Usecase represents post business logic interface.
type Usecase interface {
	Add(ctx context.Context, claims auth.Claims, np NewPost, now time.Time) (Post, error)
	Count(ctx context.Context) (int, error)
	GetAll(ctx context.Context, pageNum int, rowsPerPage int) ([]Post, error)
	GetByCatName(ctx context.Context, catName string) ([]Post, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error)
	GetByID(ctx context.Context, postID uuid.UUID) (Post, error)
	Delete(ctx context.Context, claims auth.Claims, postID uuid.UUID) error
	AddComment(ctx context.Context, claims auth.Claims, postID uuid.UUID, nc NewComment, now time.Time) (Post, error)
	GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]Comment, error)
	GetCommentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Comment, error)
	DeleteComment(ctx context.Context, claims auth.Claims, postID, commentID uuid.UUID) (Post, error)
	AddVote(ctx context.Context, clims auth.Claims, postID uuid.UUID, vote int32) (Post, error)
	GetVotesByPostID(ctx context.Context, postID uuid.UUID) ([]Vote, error)
	GetVotesByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Vote, error)
}
