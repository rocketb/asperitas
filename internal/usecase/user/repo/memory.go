package repo

import (
	"context"
	"sync"

	"github.com/rocketb/asperitas/internal/usecase/user"

	"github.com/google/uuid"
)

// Memory represents in-memory storage for users data.
type Memory struct {
	mu   *sync.RWMutex
	data map[uuid.UUID]*user.User
}

func NewMemory() *Memory {

	return &Memory{
		mu:   &sync.RWMutex{},
		data: make(map[uuid.UUID]*user.User),
	}
}

// Add Creates user in the app storage and return ID of new user.
func (r *Memory) Add(_ context.Context, user *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[user.ID] = user

	return nil
}

// GetByID Finds user by user ID in the app storage.
func (r *Memory) GetByID(_ context.Context, userID uuid.UUID) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.data[userID]
	if !ok {
		return nil, user.ErrNotFound
	}
	return u, nil
}

// GetByUsername Finds user by username in the app storage.
func (r *Memory) GetByUsername(_ context.Context, username string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.data {
		if u.Name == username {
			return u, nil
		}
	}

	return nil, user.ErrNotFound
}
