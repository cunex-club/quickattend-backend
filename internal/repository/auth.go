package repository

import (
	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type AuthRepository interface {
	GetUser(uint64) (entity.User, error)
	CreateUser(*entity.User) (*entity.User, error)
}

func (r *repository) GetUser(refId uint64) (entity.User, error) {
	var user entity.User
	err := r.db.Where(&entity.User{RefID: refId}).First(&user).Error
	return user, err
}

func (r *repository) CreateUser(user *entity.User) (*entity.User, error) {
	err := r.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, err
}
