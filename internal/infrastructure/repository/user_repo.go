package repository

import (
	"context"

	"gorm.io/gorm"

	"gagarin-soft/internal/entity"
	"gagarin-soft/internal/usecase/repo"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) repo.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Get(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
