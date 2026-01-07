package repository

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
)

type AuthRepository interface {
	GetUserById(datatypes.UUID, context.Context) (entity.User, error)
	CreateUser(*entity.User, context.Context) (*entity.User, error)
}

func (r *repository) GetUserById(userID datatypes.UUID, ctx context.Context) (entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, &entity.User{ID: userID}).Error
	return user, err
}

func (r *repository) CreateUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, err
}
