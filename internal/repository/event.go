package repository

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type EventRepository interface {
	GetParticipantRepository(ctx context.Context, refID uint64) (*entity.User, error)
}

func (r *repository) GetParticipantRepository(ctx context.Context, refID uint64) (*entity.User, error) {
	var user entity.User
	withCtx := r.db.WithContext(ctx)

	// need to check user status on the event too
	getUserErr := withCtx.Model(&user).Select("firstname_th", "surname_th", "title_th", "title_en").
		First(&user, &entity.User{RefID: refID}).Error
	if getUserErr != nil {
		return nil, getUserErr
	}
	return &user, nil
}
