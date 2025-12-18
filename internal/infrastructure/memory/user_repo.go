package memory

import (
	"context"
	"sync"

	"gagarin-soft/internal/entity"
	"gagarin-soft/internal/usecase/repo"
)

type UserRepo struct {
	mu    sync.RWMutex
	users map[string]*entity.User
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[string]*entity.User),
	}
}

var _ repo.UserRepo = (*UserRepo)(nil)

func (r *UserRepo) Get(ctx context.Context, id string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, entity.ErrUserNotFound
	}
	return user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}
