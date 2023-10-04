package repo

import (
	"database/sql"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/post"

	"github.com/google/uuid"
)

// dbPost Represents post in DB.
type dbPost struct {
	ID          uuid.UUID     `db:"post_id"`
	Type        string        `db:"type"`
	Title       string        `db:"title"`
	Category    string        `db:"category"`
	Body        string        `db:"body"`
	Score       sql.NullInt32 `db:"score"`
	Views       int           `db:"views"`
	DateCreated time.Time     `db:"date_created"`
	UserID      uuid.UUID     `db:"user_id"`
}

// dbComment Represents comment in DB.
type dbComment struct {
	ID          uuid.UUID `db:"comment_id"`
	PostID      uuid.UUID `db:"post_id"`
	UserID      uuid.UUID `db:"user_id"`
	Body        string    `db:"body"`
	DateCreated time.Time `db:"date_created"`
}

// dbVote Represents post vote in DB.
type dbVote struct {
	PostID uuid.UUID `db:"post_id"`
	UserID uuid.UUID `db:"user_id"`
	Vote   int32     `db:"vote"`
}

func toDBPost(post post.Post) dbPost {
	return dbPost{
		ID:          post.ID,
		Type:        post.Type,
		Title:       post.Title,
		Category:    post.Category,
		Body:        post.Body,
		Views:       post.Views,
		DateCreated: post.DateCreated,
		UserID:      post.UserID,
	}
}

func toCorePost(dbPost dbPost) post.Post {
	return post.Post{
		ID:          dbPost.ID,
		Type:        dbPost.Type,
		Title:       dbPost.Title,
		Category:    dbPost.Category,
		Body:        dbPost.Body,
		Score:       dbPost.Score.Int32,
		Views:       dbPost.Views,
		DateCreated: dbPost.DateCreated,
		UserID:      dbPost.UserID,
	}
}

func toCorePosts(dbPosts []dbPost) []post.Post {
	var posts []post.Post
	for _, p := range dbPosts {
		posts = append(posts, toCorePost(p))
	}

	return posts
}

func toCoreComment(dbComment dbComment) post.Comment {
	return post.Comment{
		ID:          dbComment.ID,
		PostID:      dbComment.PostID,
		UserID:      dbComment.UserID,
		Body:        dbComment.Body,
		DateCreated: dbComment.DateCreated,
	}
}

func toCoreComments(dbComments []dbComment) []post.Comment {
	var comments []post.Comment
	for _, c := range dbComments {
		comments = append(comments, toCoreComment(c))
	}

	return comments
}

func toDBComment(comment post.Comment) dbComment {
	return dbComment{
		ID:          comment.ID,
		PostID:      comment.PostID,
		UserID:      comment.UserID,
		Body:        comment.Body,
		DateCreated: comment.DateCreated,
	}
}

func toCoreVote(vote dbVote) post.Vote {
	return post.Vote{
		User:   vote.UserID,
		Vote:   vote.Vote,
		PostID: vote.PostID,
	}
}

func toCoreVotes(dbVotes []dbVote) []post.Vote {
	var votes []post.Vote
	for _, v := range dbVotes {
		votes = append(votes, toCoreVote(v))
	}

	return votes
}

func toDBVote(postID uuid.UUID, vote post.Vote) dbVote {
	return dbVote{
		PostID: postID,
		UserID: vote.User,
		Vote:   vote.Vote,
	}
}
