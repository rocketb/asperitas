package postgrp

import (
	"time"

	"github.com/rocketb/asperitas/internal/usecase/post"
	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/pkg/validate"

	"github.com/google/uuid"
)

// Author represents info about author.
type AppPostAuthor struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func toAppPostAuthor(usr user.User) AppPostAuthor {
	return AppPostAuthor{
		ID:       usr.ID.String(),
		Username: usr.Name,
	}
}

// AppPost represents application post.
type AppPost interface {
	Info()
}

// Post represents text post in storage.
type AppTextPost struct {
	ID               string        `json:"id"`
	Type             string        `json:"type"`
	Title            string        `json:"title"`
	Text             string        `json:"text"`
	Category         string        `json:"category"`
	Score            int32         `json:"score"`
	Views            int           `json:"views"`
	DateCreated      string        `json:"created"`
	Author           AppPostAuthor `json:"author"`
	UpvotePercentage int           `json:"upvotePercentage"`
	Votes            []AppVote     `json:"votes"`
	Comments         []AppComment  `json:"comments"`
}

func (p AppTextPost) Info() {}

// Post represents URL post in storage.
type AppURLPost struct {
	ID               string        `json:"id"`
	Type             string        `json:"type"`
	Title            string        `json:"title"`
	URL              string        `json:"url"`
	Category         string        `json:"category"`
	Score            int32         `json:"score"`
	Views            int           `json:"views"`
	UpvotePercentage int           `json:"upvotePercentage"`
	DateCreated      string        `json:"created"`
	Votes            []AppVote     `json:"votes"`
	Comments         []AppComment  `json:"comments"`
	Author           AppPostAuthor `json:"author"`
}

func (p AppURLPost) Info() {}

func toAppPost(p post.Post, author user.User, comments []post.Comment, commsAuthors map[uuid.UUID]user.User, votes []post.Vote) AppPost {
	switch p.Type {
	case "url":
		return AppURLPost{
			ID:               p.ID.String(),
			Type:             p.Type,
			Title:            p.Title,
			URL:              p.Body,
			Category:         p.Category,
			Score:            p.Score,
			Views:            p.Views,
			UpvotePercentage: upvotePercentage(votes),
			DateCreated:      p.DateCreated.Format(time.RFC3339),
			Author:           toAppPostAuthor(author),
			Votes:            toAppVotes(votes),
			Comments:         toAppComments(comments, commsAuthors),
		}
	default:
		return AppTextPost{
			ID:               p.ID.String(),
			Type:             p.Type,
			Title:            p.Title,
			Text:             p.Body,
			Category:         p.Category,
			Score:            p.Score,
			Views:            p.Views,
			UpvotePercentage: upvotePercentage(votes),
			DateCreated:      p.DateCreated.Format(time.RFC3339),
			Author:           toAppPostAuthor(author),
			Votes:            toAppVotes(votes),
			Comments:         toAppComments(comments, commsAuthors),
		}
	}
}

func toAppPosts(posts []post.Post, authors map[uuid.UUID]user.User, comments map[uuid.UUID][]post.Comment, commsAuthors map[uuid.UUID]user.User, votes map[uuid.UUID][]post.Vote) []AppPost {
	pss := make([]AppPost, len(posts))
	for i, p := range posts {
		pss[i] = toAppPost(p, authors[p.UserID], comments[p.ID], commsAuthors, votes[p.ID])
	}
	return pss
}

// Comment represents info about post comments.
type AppComment struct {
	ID          string        `json:"id"`
	PostID      string        `json:"-"`
	DateCreated string        `json:"created"`
	Author      AppPostAuthor `json:"author"`
	Body        string        `json:"body"`
}

func toAppComment(comment post.Comment, author user.User) AppComment {
	return AppComment{
		ID:          comment.ID.String(),
		PostID:      comment.PostID.String(),
		DateCreated: comment.DateCreated.Format(time.RFC3339),
		Author:      toAppPostAuthor(author),
		Body:        comment.Body,
	}
}

func toAppComments(comments []post.Comment, authors map[uuid.UUID]user.User) []AppComment {
	comms := make([]AppComment, len(comments))
	for i, c := range comments {
		comms[i] = toAppComment(c, authors[c.UserID])
	}

	return comms
}

// Vote represents info about post votes.
type AppVote struct {
	Vote int32  `json:"vote"`
	User string `json:"user"`
}

func toAppVote(vote post.Vote) AppVote {
	return AppVote{
		Vote: vote.Vote,
		User: vote.User.String(),
	}
}

func toAppVotes(votes []post.Vote) []AppVote {
	vts := make([]AppVote, len(votes))
	for i, v := range votes {
		vts[i] = toAppVote(v)
	}

	return vts
}

// NewPost is what we require from user to add a Post.
type AppNewPost struct {
	Title    string `json:"title" validate:"required"`
	Type     string `json:"type" default:"text" validate:"required,oneof=text url"`
	Text     string `json:"text" validate:"required_if=Type text"`
	URL      string `json:"url" validate:"required_if=Type url"`
	Category string `json:"category" validate:"required"`
}

func toCoreNewPost(np AppNewPost) post.NewPost {
	return post.NewPost{
		Title:    np.Title,
		Type:     np.Type,
		Text:     np.Text,
		URL:      np.URL,
		Category: np.Category,
	}
}

// Validate checks the data in the model is considered clean.
func (app AppNewPost) Validate() error {
	return validate.Check(app)
}

// NewComment is what we require from user to add a Comment.
type AppNewComment struct {
	Text string `json:"text" validate:"required"`
}

// Validate checks the data in the model is considered clean.
func (app AppNewComment) Validate() error {
	return validate.Check(app)
}

func toCoreNewComment(nc AppNewComment) post.NewComment {
	return post.NewComment{
		Text: nc.Text,
	}
}

// upvotePercentage count post upvote percentage.
func upvotePercentage(votes []post.Vote) int {
	if len(votes) == 0 {
		return 0
	}

	var aye float32
	for _, v := range votes {
		if v.Vote > 0 {
			aye++
		}
	}

	return int(aye / float32(len(votes)) * 100)
}
