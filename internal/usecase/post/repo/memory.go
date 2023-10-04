package repo

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/rocketb/asperitas/internal/usecase/post"
)

// PostDB represents single post in the storage.
type PostDB struct {
	data     *post.Post
	comments map[uuid.UUID]*post.Comment
	votes    []*post.Vote
}

// CommentsList Returns all comments as list.
func (p PostDB) CommentsList() []*post.Comment {
	comms := make([]*post.Comment, 0, len(p.comments))
	for _, c := range p.comments {
		comms = append(comms, c)
	}
	return comms
}

// Memory Represents in-memory storage for posts data.
type Memory struct {
	mu    *sync.RWMutex
	posts map[uuid.UUID]*PostDB
}

// NewMemory Creates new in-memory storage.
func NewMemory() *Memory {
	return &Memory{
		mu:    &sync.RWMutex{},
		posts: make(map[uuid.UUID]*PostDB),
	}
}

// GetAll Return all posts from the app storage.
func (r *Memory) GetAll(_ context.Context) ([]*post.Post, error) {
	posts := make([]*post.Post, 0, len(r.posts))

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.posts {
		posts = append(posts, p.data)
	}

	return posts, nil
}

// GetByCatName Finds posts of given category.
func (r *Memory) GetByCatName(_ context.Context, catName string) ([]*post.Post, error) {
	posts := []*post.Post{}
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.posts {
		if p.data.Category != catName {
			continue
		}
		posts = append(posts, p.data)
	}
	return posts, nil
}

// GetByUserID Finds posts of given user by user ID.
func (r *Memory) GetByUserID(_ context.Context, userID uuid.UUID) ([]*post.Post, error) {
	posts := []*post.Post{}
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.posts {
		if p.data.UserID != userID {
			continue
		}
		posts = append(posts, p.data)
	}
	return posts, nil
}

// GetByID Finds post by it ID.
func (r *Memory) GetByID(_ context.Context, postID uuid.UUID) (*post.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.posts[postID]
	if !ok {
		return nil, post.ErrNotFound
	}
	return p.data, nil
}

// Add Creates post in the app storage.
func (r *Memory) Add(_ context.Context, newPost *post.Post) (postID uuid.UUID, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.posts[newPost.ID] = &PostDB{
		data:     newPost,
		comments: make(map[uuid.UUID]*post.Comment),
		votes: []*post.Vote{
			{Vote: 1, User: newPost.UserID},
		},
	}
	return newPost.ID, nil
}

// Delete Removes post from the app storage.
func (r *Memory) Delete(_ context.Context, postID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.posts, postID)
	return nil
}

// GetVotes Finds votes of the given post by ID in the app storage.
func (r *Memory) GetVotes(_ context.Context, postID uuid.UUID) ([]*post.Vote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.posts[postID]
	if !ok {
		return nil, post.ErrNotFound
	}

	return p.votes, nil
}

// AddVote Adds vote to the givven post.
func (r *Memory) AddVote(_ context.Context, postID uuid.UUID, vote *post.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.posts[postID]
	if !ok {
		return post.ErrNotFound
	}

	var found bool
	for _, v := range p.votes {
		if v.User == vote.User {
			v.Vote = vote.Vote
			found = true
			break
		}
	}
	if !found {
		p.votes = append(p.votes, vote)
	}

	p.data.Score += vote.Vote

	return nil
}

// AddComment Adds comment to the givven post.
func (r *Memory) AddComment(_ context.Context, postID uuid.UUID, newComment *post.Comment) (commentID uuid.UUID, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.posts[postID]; !ok {
		return uuid.UUID{}, post.ErrNotFound
	}

	r.posts[postID].comments[newComment.ID] = newComment

	return newComment.ID, nil
}

// GetComments Finds comments of the givven post in the app strorage.
func (r *Memory) GetComments(_ context.Context, postID uuid.UUID) ([]*post.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.posts[postID]
	if !ok {
		return nil, post.ErrNotFound
	}

	return p.CommentsList(), nil
}

// GetCommentByID Finds comment by comment ID.
func (r *Memory) GetCommentByID(_ context.Context, commentID uuid.UUID) (*post.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.posts {
		c, ok := p.comments[commentID]
		if ok {
			return c, nil
		}
	}

	return nil, post.ErrCommentNotFound
}

// DeleteComment Removes comment from storage by post, comment IDs.
func (r *Memory) DeleteComment(_ context.Context, postID, commentID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.posts[postID]; !ok {
		return post.ErrNotFound
	}

	delete(r.posts[postID].comments, commentID)

	return nil
}
