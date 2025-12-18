package usecase

import (
	"context"
	"fmt"

	"gagarin-soft/internal/entity"
	"gagarin-soft/internal/usecase/repo"
)

// UserUseCase is the service layer for Users.
type UserUseCase struct {
	repo repo.UserRepo
}

// NewUserUseCase creates a new UserUseCase.
func NewUserUseCase(r repo.UserRepo) *UserUseCase {
	return &UserUseCase{
		repo: r,
	}
}

// GetUser retrieves a user by ID.
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*entity.User, error) {
	return uc.repo.Get(ctx, id)
}

// CreateUser creates a new user.
func (uc *UserUseCase) CreateUser(ctx context.Context, user *entity.User) error {
	if user.Name == "" {
		return fmt.Errorf("user name cannot be empty")
	}
	return uc.repo.Create(ctx, user)
}
