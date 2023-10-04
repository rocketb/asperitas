package repo

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/rocketb/asperitas/internal/usecase/post"
	db "github.com/rocketb/asperitas/pkg/database/pgx"
	"github.com/rocketb/asperitas/pkg/database/pgx/dbarray"
	"github.com/rocketb/asperitas/pkg/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Postgres represents postgres storage for posts data.
type Postgres struct {
	db  *sqlx.DB
	log *logger.Logger
}

func NewPostgres(db *sqlx.DB, log *logger.Logger) *Postgres {
	return &Postgres{
		db:  db,
		log: log,
	}
}

// GetAll return all posts from the app storage.
func (r *Postgres) GetAll(ctx context.Context, pageNum int, rowsPerPage int) ([]post.Post, error) {
	data := map[string]interface{}{
		"offset":        (pageNum - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `
	SELECT
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id, SUM(v.vote) as score
	FROM
		posts p
	LEFT JOIN
		votes v ON p.post_id = v.post_id
	GROUP BY
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id
	`

	buf := bytes.NewBufferString(q)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var posts []dbPost
	if err := db.NamedQuerySlice(ctx, r.log, r.db, buf.String(), data, &posts); err != nil {
		return nil, fmt.Errorf("selecting all posts: %w", err)
	}

	return toCorePosts(posts), nil
}

// GetByUserID finds posts of given user by user ID.
func (r *Postgres) GetByUserID(ctx context.Context, userID uuid.UUID) ([]post.Post, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID.String(),
	}
	const q = `
	SELECT
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id, SUM(v.vote) as score
	FROM
		posts p
	LEFT JOIN
		votes v ON p.post_id = v.post_id
	WHERE
		p.user_id = :user_id
	GROUP BY
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id
	`

	var posts []dbPost
	if err := db.NamedQuerySlice(ctx, r.log, r.db, q, data, &posts); err != nil {
		return nil, fmt.Errorf("selecting posts by user_id(%s): %w", userID, err)
	}

	return toCorePosts(posts), nil
}

// GetByCatName finds posts of given category.
func (r *Postgres) GetByCatName(ctx context.Context, catName string) ([]post.Post, error) {
	data := struct {
		Category string `db:"category"`
	}{
		Category: catName,
	}
	const q = `
	SELECT
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id, SUM(v.vote) as score
	FROM
		posts p
	LEFT JOIN
		votes v ON p.post_id = v.post_id
	WHERE
		p.category = :category
	GROUP BY
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id
	`

	var posts []dbPost
	if err := db.NamedQuerySlice(ctx, r.log, r.db, q, data, &posts); err != nil {
		return nil, fmt.Errorf("selecting posts by category(%s): %w", catName, err)
	}

	return toCorePosts(posts), nil
}

// GetByID finds post by it ID.
func (r *Postgres) GetByID(ctx context.Context, postID uuid.UUID) (post.Post, error) {
	data := struct {
		PostID string `db:"post_id"`
	}{
		PostID: postID.String(),
	}
	const q = `
	SELECT
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id, SUM(v.vote) as score
	FROM
		posts p
	LEFT JOIN
		votes v ON p.post_id = v.post_id
	WHERE
		p.post_id = :post_id
	GROUP BY
		p.post_id, p.type, p.title, p.category, p.body, p.views, p.date_created, p.user_id
	`

	var p dbPost
	if err := db.NamedQueryStruct(ctx, r.log, r.db, q, data, &p); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return post.Post{}, post.ErrNotFound
		}
		return post.Post{}, fmt.Errorf("selecting post by post_id: %w", err)
	}

	return toCorePost(p), nil
}

// Add create post in the app storage.
func (r *Postgres) Add(ctx context.Context, newPost post.Post) error {
	const q = `
	INSERT INTO posts
		(post_id, type, title, category, body, views, date_created, user_id)
	VALUES
		(:post_id, :type, :title, :category, :body, :views, :date_created, :user_id)
	`

	if err := db.NamedExecContext(ctx, r.log, r.db, q, toDBPost(newPost)); err != nil {
		return fmt.Errorf("adding post: %w", err)
	}

	return nil
}

// Delete Removes post from the app storage.
func (r *Postgres) Delete(ctx context.Context, postID uuid.UUID) error {
	data := struct {
		PostID string `db:"post_id"`
	}{
		PostID: postID.String(),
	}
	const q = `
	DELETE FROM
		posts
	WHERE
		post_id = :post_id`

	if err := db.NamedExecContext(ctx, r.log, r.db, q, data); err != nil {
		return fmt.Errorf("deleting post(%s): %w", postID, err)
	}

	return nil
}

// Count retunns total number of posts in the DB.
func (r *Postgres) Count(ctx context.Context) (int, error) {
	const q = `
	SELECT
		count(1)
	FROM
		posts
	`

	var count struct {
		Count int `db:"count"`
	}

	if err := db.QueryStruct(ctx, r.log, r.db, q, &count); err != nil {
		return 0, fmt.Errorf("quering total posts count: %w", err)
	}

	return count.Count, nil
}

// GetComments returns a list of post comments.
func (r *Postgres) GetCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]post.Comment, error) {
	data := struct {
		PostID string `db:"post_id"`
	}{
		PostID: postID.String(),
	}
	const q = `
	SELECT
		comment_id, post_id, date_created, body, user_id
	FROM
		comments
	WHERE
		post_id = :post_id
	`

	var comments []dbComment
	if err := db.NamedQuerySlice(ctx, r.log, r.db, q, data, &comments); err != nil {
		return nil, fmt.Errorf("selecting post(%s) comments: %w", postID, err)
	}

	return toCoreComments(comments), nil
}

// GetCommentsByPostIDs returns a list of posts comments.
func (r *Postgres) GetCommentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]post.Comment, error) {
	ids := make([]string, len(postIDs))
	for i, pid := range postIDs {
		ids[i] = pid.String()
	}

	data := struct {
		PostID interface {
			driver.Valuer
			sql.Scanner
		} `db:"post_id"`
	}{
		PostID: dbarray.Array(ids),
	}

	const q = `
	SELECT
		comment_id, post_id, date_created, body, user_id
	FROM
		comments
	WHERE
		post_id = ANY(:post_id)
	`

	var comments []dbComment
	if err := db.NamedQuerySlice(ctx, r.log, r.db, q, data, &comments); err != nil {
		return nil, fmt.Errorf("selecting posts comments: %w", err)
	}

	return toCoreComments(comments), nil
}

// GetCommentByID finds comment by it ID.
func (r *Postgres) GetCommentByID(ctx context.Context, commentID uuid.UUID) (post.Comment, error) {
	data := struct {
		CommentID string `db:"comment_id"`
	}{
		CommentID: commentID.String(),
	}
	const q = `
	SELECT comment_id, date_created, body, user_id
	FROM
		comments
	WHERE
		comment_id = :comment_id
	`

	var comment dbComment
	if err := db.NamedQueryStruct(ctx, r.log, r.db, q, data, &comment); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return post.Comment{}, post.ErrCommentNotFound
		}
		return post.Comment{}, fmt.Errorf("selecting comments: %w", err)
	}

	return toCoreComment(comment), nil
}

// AddComment create comment in the app storage.
func (r *Postgres) AddComment(ctx context.Context, newComment post.Comment) error {
	const q = `
	INSERT INTO comments
		(comment_id, post_id, user_id, body, date_created)
	VALUES
		(:comment_id, :post_id, :user_id, :body, :date_created)
	`
	if err := db.NamedExecContext(ctx, r.log, r.db, q, toDBComment(newComment)); err != nil {
		return fmt.Errorf("adding comment: %w", err)
	}

	return nil
}

// DeleteComment removes comment from the app storage.
func (r *Postgres) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	data := struct {
		CommentID string `db:"comment_id"`
	}{
		CommentID: commentID.String(),
	}
	const q = `
	DELETE FROM
		comments
	WHERE
		comment_id = :comment_id
	`

	if err := db.NamedExecContext(ctx, r.log, r.db, q, data); err != nil {
		return fmt.Errorf("deleting comment(%s): %w", commentID, err)
	}

	return nil
}

// GetVotes returns a list of post votes.
func (r *Postgres) GetVotesByPostID(ctx context.Context, postID uuid.UUID) ([]post.Vote, error) {
	data := struct {
		PostID string `db:"post_id"`
	}{
		PostID: postID.String(),
	}
	const q = `
	SELECT
		post_id, user_id, vote
	FROM
		votes
	WHERE
		post_id = :post_id
	`

	var votes []dbVote
	if err := db.NamedQuerySlice(ctx, r.log, r.db, q, data, &votes); err != nil {
		return nil, fmt.Errorf("getting post(%s) votes: %w", postID, err)
	}

	return toCoreVotes(votes), nil
}

// GetVotesByPostIDs returns a list of posts votes.
func (r *Postgres) GetVotesByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]post.Vote, error) {
	ids := make([]string, len(postIDs))
	for i, pid := range postIDs {
		ids[i] = pid.String()
	}

	data := struct {
		PostID interface {
			driver.Valuer
			sql.Scanner
		} `db:"post_id"`
	}{
		PostID: dbarray.Array(ids),
	}

	const q = `
	SELECT
		post_id, user_id, vote
	FROM
		votes
	WHERE
		post_id = ANY(:post_id)
	`

	var votes []dbVote
	if err := db.NamedQuerySlice(ctx, r.log, r.db, q, data, &votes); err != nil {
		return nil, fmt.Errorf("getting posts votes: %w", err)
	}

	return toCoreVotes(votes), nil
}

// AddVote creates vote in the storage.
func (r *Postgres) AddVote(ctx context.Context, postID uuid.UUID, vote post.Vote) error {
	const q = `
	INSERT INTO votes
		(post_id, user_id, vote)
	VALUES
		(:post_id, :user_id, :vote)
	`

	if err := db.NamedExecContext(ctx, r.log, r.db, q, toDBVote(postID, vote)); err != nil {
		return fmt.Errorf("adding vote: %w", err)
	}

	return nil
}

// UpdateVote changes vote of the user in the storage.
func (r *Postgres) UpdateVote(ctx context.Context, postID uuid.UUID, vote post.Vote) error {
	data := struct {
		PostID string `db:"post_id"`
		UserID string `db:"user_id"`
		Vote   int32  `db:"vote"`
	}{
		PostID: postID.String(),
		UserID: vote.User.String(),
		Vote:   vote.Vote,
	}
	const q = `
	UPDATE
		votes
	SET
		vote = :vote
	WHERE
		post_id = :post_id and user_id = :user_id
	`

	if err := db.NamedExecContext(ctx, r.log, r.db, q, data); err != nil {
		return fmt.Errorf("updating vote: %w", err)
	}

	return nil
}

// CheckVote checks if user already voted or not.
func (r *Postgres) CheckVote(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	data := struct {
		PostID string `db:"post_id"`
		UserID string `db:"user_id"`
	}{
		PostID: postID.String(),
		UserID: userID.String(),
	}
	const q = `
	SELECT COUNT (vote)
	FROM
		votes
	WHERE
		post_id = :post_id and user_id = :user_id
	`
	v := struct {
		Count int `db:"count"`
	}{}
	if err := db.NamedQueryStruct(ctx, r.log, r.db, q, data, &v); err != nil {
		return fmt.Errorf("checking if user vote exist: %w", err)
	}

	if v.Count == 0 {
		return post.ErrNotFound
	}

	return nil
}
