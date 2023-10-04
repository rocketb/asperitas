package post

import (
	"context"
	"errors"
	"time"

	"github.com/rocketb/asperitas/internal/web/auth"

	"github.com/google/uuid"
)

var (
	ErrNotFound        = errors.New("post not found")
	ErrWrongPostType   = errors.New("new post should be url or text")
	ErrForbidden       = errors.New("action is not allowed")
	ErrCommentNotFound = errors.New("comment not found")
)

type Core struct {
	PostsRepo Repo
	idGen     func() uuid.UUID
}

func NewCore(postsRepo Repo) *Core {
	return &Core{
		idGen:     uuid.New,
		PostsRepo: postsRepo,
	}
}

// GetAll gets all posts.
func (u *Core) GetAll(ctx context.Context, pageNum int, rowsPerPage int) ([]Post, error) {
	posts, err := u.PostsRepo.GetAll(ctx, pageNum, rowsPerPage)
	if err != nil {
		return []Post{}, err
	}

	return posts, nil
}

// Count returns total number of posts.
func (u *Core) Count(ctx context.Context) (int, error) {
	total, err := u.PostsRepo.Count(ctx)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetByCatName finds all posts of category.
func (u *Core) GetByCatName(ctx context.Context, catName string) ([]Post, error) {
	posts, err := u.PostsRepo.GetByCatName(ctx, catName)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetByUsername finds posts of user by user ID.
func (u *Core) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error) {
	posts, err := u.PostsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetByID gets post by it ID.
func (u *Core) GetByID(ctx context.Context, postID uuid.UUID) (Post, error) {
	p, err := u.PostsRepo.GetByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	return p, nil
}

// Add creates a post.
func (u *Core) Add(ctx context.Context, claims auth.Claims, np NewPost, now time.Time) (Post, error) {
	if np.Type != "url" && np.Type != "text" {
		return Post{}, ErrWrongPostType
	}

	body := np.Text
	if np.Type == "url" {
		body = np.URL
	}

	p := Post{
		ID:          u.idGen(),
		Type:        np.Type,
		Title:       np.Title,
		Category:    np.Category,
		Body:        body,
		Score:       1,
		Views:       0,
		DateCreated: now,
		UserID:      claims.User.ID,
	}

	if err := u.PostsRepo.Add(ctx, p); err != nil {
		return Post{}, err
	}

	if err := u.PostsRepo.AddVote(ctx, p.ID, Vote{Vote: 1, User: p.UserID}); err != nil {
		return Post{}, err
	}

	return p, nil
}

// Delete removes the post identified by given post ID.
func (u *Core) Delete(ctx context.Context, claims auth.Claims, postID uuid.UUID) error {
	p, err := u.PostsRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	if p.UserID != claims.User.ID {
		return ErrForbidden
	}

	return u.PostsRepo.Delete(ctx, postID)
}

// AddVote addds vote(upvote/downvote) to the givven post by post ID.
func (u *Core) AddVote(ctx context.Context, claims auth.Claims, postID uuid.UUID, vote int32) (Post, error) {
	if _, err := u.PostsRepo.GetByID(ctx, postID); err != nil {
		return Post{}, err
	}

	newVote := Vote{
		Vote: vote,
		User: claims.User.ID,
	}
	if err := u.PostsRepo.CheckVote(ctx, postID, claims.User.ID); err != nil {
		if err != ErrNotFound {
			return Post{}, err
		}

		if err := u.PostsRepo.AddVote(ctx, postID, newVote); err != nil {
			return Post{}, err
		}
	} else {
		if err := u.PostsRepo.UpdateVote(ctx, postID, newVote); err != nil {
			return Post{}, err
		}
	}

	p, err := u.PostsRepo.GetByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	return p, nil
}

// GetVotesByPostID finds votes by post ID.
func (u *Core) GetVotesByPostID(ctx context.Context, postID uuid.UUID) ([]Vote, error) {
	votes, err := u.PostsRepo.GetVotesByPostID(ctx, postID)
	if err != nil {
		return []Vote{}, err
	}

	return votes, nil
}

// GetVotesByPostIDs finds posts votes by post IDs.
func (u *Core) GetVotesByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Vote, error) {
	votes, err := u.PostsRepo.GetVotesByPostIDs(ctx, postIDs)
	if err != nil {
		return []Vote{}, err
	}

	return votes, nil
}

// AddComment adds comment to the given post by post ID.
func (u *Core) AddComment(ctx context.Context, claims auth.Claims, postID uuid.UUID, nc NewComment, now time.Time) (Post, error) {
	if _, err := u.PostsRepo.GetByID(ctx, postID); err != nil {
		return Post{}, err
	}

	comment := Comment{
		ID:          u.idGen(),
		PostID:      postID,
		DateCreated: now,
		UserID:      claims.User.ID,
		Body:        nc.Text,
	}

	if err := u.PostsRepo.AddComment(ctx, comment); err != nil {
		return Post{}, err
	}

	p, err := u.PostsRepo.GetByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	return p, nil
}

// GetCommentsByPostID finds post comments by post ID.
func (u *Core) GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]Comment, error) {
	comments, err := u.PostsRepo.GetCommentsByPostID(ctx, postID)
	if err != nil {
		return []Comment{}, err
	}

	return comments, nil
}

// GetCommentsByPostIDs finds posts comments by post IDs.
func (u *Core) GetCommentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Comment, error) {
	comments, err := u.PostsRepo.GetCommentsByPostIDs(ctx, postIDs)
	if err != nil {
		return []Comment{}, err
	}

	return comments, nil
}

// DeleteComment deletes comments of the given post by post and comment IDs.
func (u *Core) DeleteComment(ctx context.Context, claims auth.Claims, postID, commentID uuid.UUID) (Post, error) {
	if _, err := u.PostsRepo.GetByID(ctx, postID); err != nil {
		return Post{}, err
	}

	comment, err := u.PostsRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return Post{}, err
	}

	if comment.UserID != claims.User.ID {
		return Post{}, ErrForbidden
	}

	if err = u.PostsRepo.DeleteComment(ctx, commentID); err != nil {
		return Post{}, err
	}

	p, err := u.PostsRepo.GetByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	return p, nil
}
