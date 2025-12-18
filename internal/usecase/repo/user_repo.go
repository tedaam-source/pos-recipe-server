package repo

import (
	"context"

	"gagarin-soft/internal/entity"
)

// UserRepo defines the interface for storing and retrieving users.
type UserRepo interface {
	Get(ctx context.Context, id string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
}
