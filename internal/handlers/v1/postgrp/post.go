package postgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/post"
	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/internal/web/auth"
	"github.com/rocketb/asperitas/internal/web/paging"
	"github.com/rocketb/asperitas/internal/web/request"
	"github.com/rocketb/asperitas/pkg/validate"
	"github.com/rocketb/asperitas/pkg/web"

	"github.com/google/uuid"
)

type PostsHandler struct {
	Posts post.Usecase
	Users user.Usecase
}

// List return a list of posts.
func (h *PostsHandler) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := paging.ParseRequest(r)
	if err != nil {
		return err
	}

	pss, err := h.Posts.GetAll(ctx, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("collecting posts: %w", err)
	}

	appPosts, err := h.getPostsInfo(ctx, pss)
	if err != nil {
		return err
	}

	total, err := h.Posts.Count(ctx)
	if err != nil {
		return fmt.Errorf("counting posts: %w", err)
	}

	return web.Respond(ctx, w, paging.NewResponse(appPosts, total, page.Number, page.RowsPerPage), http.StatusOK)
}

// ListByCatName returns a list of post of given category.
func (h *PostsHandler) ListByCatName(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := paging.ParseRequest(r)
	if err != nil {
		return err
	}

	pss, err := h.Posts.GetByCatName(ctx, web.Param(r, "category_name"))
	if err != nil {
		return fmt.Errorf("collecting posts by category: %w", err)
	}

	appPosts, err := h.getPostsInfo(ctx, pss)
	if err != nil {
		return err
	}

	total, err := h.Posts.Count(ctx)
	if err != nil {
		return fmt.Errorf("counting posts: %w", err)
	}

	return web.Respond(ctx, w, paging.NewResponse(appPosts, total, page.Number, page.RowsPerPage), http.StatusOK)
}

// ListByUsername returns a list of posts of given user.
func (h *PostsHandler) ListByUsername(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := paging.ParseRequest(r)
	if err != nil {
		return err
	}

	usr, err := h.Users.GetByUsername(ctx, web.Param(r, "user_name"))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return request.NewError(err, http.StatusBadRequest)
		}
		return fmt.Errorf("getting user: %w", err)
	}

	pss, err := h.Posts.GetByUserID(ctx, usr.ID)
	if err != nil {
		return fmt.Errorf("collecting posts by username: %w", err)
	}

	appPosts, err := h.getPostsInfo(ctx, pss)
	if err != nil {
		return err
	}

	total, err := h.Posts.Count(ctx)
	if err != nil {
		return fmt.Errorf("counting posts: %w", err)
	}

	return web.Respond(ctx, w, paging.NewResponse(appPosts, total, page.Number, page.RowsPerPage), http.StatusOK)
}

// GetByID returns a post by its ID.
func (h *PostsHandler) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pid, err := uuid.Parse(web.Param(r, "post_id"))
	if err != nil {
		return validate.NewFieldsError("post_id", err)
	}

	p, err := h.Posts.GetByID(ctx, pid)
	if err != nil {
		switch {
		case errors.Is(err, post.ErrNotFound):
			return request.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("getting post(%s): %w", pid, err)
		}
	}

	appPost, err := h.getPostInfo(ctx, p)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, appPost, http.StatusOK)
}

// AddPost adds a new post to the app.
func (h *PostsHandler) AddPost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var np AppNewPost
	if err := web.Decode(r, &np); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	p, err := h.Posts.Add(ctx, auth.GetClaims(ctx), toCoreNewPost(np), time.Now())
	if err != nil {
		switch err {
		case post.ErrWrongPostType:
			return request.NewError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("creating new post: %w", err)
		}
	}

	appPost, err := h.getPostInfo(ctx, p)
	if err != nil {
		return err
	}
	return web.Respond(ctx, w, appPost, http.StatusCreated)
}

// DeleteByID deletes given post by its ID.
func (h *PostsHandler) DeleteByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pid, err := uuid.Parse(web.Param(r, "post_id"))
	if err != nil {
		return validate.NewFieldsError("post_id", err)
	}

	err = h.Posts.Delete(ctx, auth.GetClaims(ctx), pid)
	if err != nil {
		switch err {
		case post.ErrForbidden:
			return request.NewError(err, http.StatusForbidden)
		case post.ErrNotFound:
			return request.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("deleting post(%s): %w", pid, err)
		}
	}
	return web.Respond(ctx, w, web.MessageResponse{Msg: "success"}, http.StatusOK)
}

// AddComment adds comment for a given post.
func (h *PostsHandler) AddComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var nc AppNewComment
	if err := web.Decode(r, &nc); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	pid, err := uuid.Parse(web.Param(r, "post_id"))
	if err != nil {
		return validate.NewFieldsError("post_id", err)
	}

	p, err := h.Posts.AddComment(ctx, auth.GetClaims(ctx), pid, toCoreNewComment(nc), time.Now())
	if err != nil {
		switch err {
		case post.ErrNotFound:
			return request.NewError(post.ErrNotFound, http.StatusBadRequest)
		default:
			return fmt.Errorf("creating comment for post(%s): %w", pid, err)
		}
	}

	appPost, err := h.getPostInfo(ctx, p)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, appPost, http.StatusCreated)
}

// DeleteComment deletes comment by post and comment IDs.
func (h *PostsHandler) DeleteComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	cid, err := uuid.Parse(web.Param(r, "comment_id"))
	if err != nil {
		return validate.NewFieldsError("comment_id", err)
	}

	pid, err := uuid.Parse(web.Param(r, "post_id"))
	if err != nil {
		return validate.NewFieldsError("post_id", err)
	}

	p, err := h.Posts.DeleteComment(ctx, auth.GetClaims(ctx), pid, cid)
	if err != nil {
		switch err {
		case post.ErrForbidden:
			return request.NewError(err, http.StatusForbidden)
		case post.ErrNotFound:
			return request.NewError(err, http.StatusNotFound)
		case post.ErrCommentNotFound:
			return request.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("deleting comment(%s) of post(%s): %w", cid, pid, err)
		}
	}

	appPost, err := h.getPostInfo(ctx, p)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, appPost, http.StatusOK)
}

// UpVote add vote to the given post.
func (h *PostsHandler) UpVote(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pid, err := uuid.Parse(web.Param(r, "post_id"))
	if err != nil {
		return validate.NewFieldsError("post_id", err)
	}

	p, err := h.Posts.AddVote(ctx, auth.GetClaims(ctx), pid, 1)
	if err != nil {
		switch err {
		case post.ErrNotFound:
			return request.NewError(post.ErrNotFound, http.StatusBadRequest)
		default:
			return fmt.Errorf("upvote post(%s): %w", pid, err)
		}
	}

	appPost, err := h.getPostInfo(ctx, p)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, appPost, http.StatusOK)
}

// DownVote removes vote or downvote post.
func (h *PostsHandler) DownVote(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pid, err := uuid.Parse(web.Param(r, "post_id"))
	if err != nil {
		return validate.NewFieldsError("post_id", err)
	}

	p, err := h.Posts.AddVote(ctx, auth.GetClaims(ctx), pid, -1)
	if err != nil {
		switch err {
		case post.ErrNotFound:
			return request.NewError(post.ErrNotFound, http.StatusBadRequest)
		default:
			return fmt.Errorf("downvote post(%s): %w", pid, err)
		}
	}

	appPost, err := h.getPostInfo(ctx, p)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, appPost, http.StatusOK)
}

// getPostsInfo collects extended posts info, including users, comments and
// votes data.
func (h *PostsHandler) getPostsInfo(ctx context.Context, pss []post.Post) ([]AppPost, error) {
	pAuthors := make(map[uuid.UUID]user.User)
	comments := make(map[uuid.UUID][]post.Comment)
	cAuthors := make(map[uuid.UUID]user.User)
	votes := make(map[uuid.UUID][]post.Vote)

	if len(pss) > 0 {
		postIDs := make([]uuid.UUID, 0, len(pss))
		for _, p := range pss {
			postIDs = append(postIDs, p.ID)
			pAuthors[p.UserID] = user.User{}
			comments[p.ID] = []post.Comment{}
			votes[p.ID] = []post.Vote{}
		}

		userIDs := make([]uuid.UUID, 0, len(pAuthors))
		for uid := range pAuthors {
			userIDs = append(userIDs, uid)
		}
		usrs, err := h.Users.GetByIDs(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("collecting users: %w", err)
		}

		for _, u := range usrs {
			pAuthors[u.ID] = u
		}

		// ------- collect comments and comments authors
		comms, err := h.Posts.GetCommentsByPostIDs(ctx, postIDs)
		if err != nil {
			return nil, fmt.Errorf("collecting comments: %w", err)
		}

		for _, c := range comms {
			comments[c.PostID] = append(comments[c.PostID], c)
			cAuthors[c.UserID] = user.User{}
		}

		commsUsrsIDs := make([]uuid.UUID, 0, len(cAuthors))
		for uid := range cAuthors {
			commsUsrsIDs = append(commsUsrsIDs, uid)
		}
		commUsrs, err := h.Users.GetByIDs(ctx, commsUsrsIDs)
		if err != nil {
			return nil, fmt.Errorf("collecting comment authors: %w", err)
		}

		for _, u := range commUsrs {
			cAuthors[u.ID] = u
		}

		// // ------- collect votes
		vts, err := h.Posts.GetVotesByPostIDs(ctx, postIDs)
		if err != nil {
			return nil, fmt.Errorf("collecting votes: %w", err)
		}

		for _, v := range vts {
			votes[v.PostID] = append(votes[v.PostID], v)
		}
	}

	return toAppPosts(pss, pAuthors, comments, cAuthors, votes), nil
}

// getPostInfo collects extended post info, including users, comments and
// votes data.
func (h *PostsHandler) getPostInfo(ctx context.Context, p post.Post) (AppPost, error) {
	author, err := h.Users.GetByID(ctx, p.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting post author: %w", err)
	}

	comments, err := h.Posts.GetCommentsByPostID(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("getting post comments: %w", err)
	}

	commentsAuthors := make(map[uuid.UUID]user.User)
	if len(comments) > 0 {
		for _, c := range comments {
			commentsAuthors[c.UserID] = user.User{}
		}

		authorsIDs := make([]uuid.UUID, 0, len(commentsAuthors))
		for _, a := range comments {
			authorsIDs = append(authorsIDs, a.UserID)
		}

		authors, err := h.Users.GetByIDs(ctx, authorsIDs)
		if err != nil {
			return nil, fmt.Errorf("getting post comments authors: %w", err)
		}

		for _, a := range authors {
			commentsAuthors[a.ID] = a
		}
	}

	votes, err := h.Posts.GetVotesByPostID(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("getting post votes: %w", err)
	}

	return toAppPost(p, author, comments, commentsAuthors, votes), nil
}
