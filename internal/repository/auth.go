package repository

import (
	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type AuthRepository interface {
	GetUser(uint64) (entity.User, error)
}

func (r *repository) GetUser(refId uint64) (entity.User, error) {
	var user entity.User
	err := r.db.Where(&entity.User{RefID: refId}).First(&user).Error
	return user, err
}
