package repository

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type AuthRepository interface {
	GetUserByRefId(uint64, context.Context) (entity.User, error)
	CreateUser(*entity.User, context.Context) (*entity.User, error)
}

func (r *repository) GetUserByRefId(refId uint64, ctx context.Context) (entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where(&entity.User{RefID: refId}).First(&user).Error
	return user, err
}

func (r *repository) CreateUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, err
}
