package repository

import (
	"context"
	"fmt"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthRepository interface {
	GetUserByRefId(uint64, context.Context) (entity.User, error)
	CreateUser(*entity.User, context.Context) (*entity.User, error)

	FindWhitelistPendingByRefID(ctx context.Context, refID uint64) ([]entity.EventWhitelistPending, error)
	DeleteWhitelistPendingByRefID(ctx context.Context, refID uint64) error
	SyncWhitelistPendingToWhitelist(ctx context.Context, refID uint64) error
}

func (r *repository) GetUserByRefId(refId uint64, ctx context.Context) (entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, &entity.User{RefID: refId}).Error
	return user, err
}

func (r *repository) CreateUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, err
}

func (r *repository) FindWhitelistPendingByRefID(ctx context.Context, refID uint64) ([]entity.EventWhitelistPending, error) {
	var pend []entity.EventWhitelistPending
	err := r.db.WithContext(ctx).Where("attendee_ref_id = ?", refID).Find(&pend).Error
	return pend, err
}

func (r *repository) DeleteWhitelistPendingByRefID(ctx context.Context, refID uint64) error {
	return r.db.WithContext(ctx).Where("attendee_ref_id = ?", refID).Delete(&entity.EventWhitelistPending{}).Error
}

func (r *repository) SyncWhitelistPendingToWhitelist(ctx context.Context, refID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pend []entity.EventWhitelistPending
		if err := tx.Where("attendee_ref_id = ?", refID).Find(&pend).Error; err != nil {
			return err
		}
		if len(pend) == 0 {
			return nil
		}

		wl := make([]entity.EventWhitelist, 0, len(pend))
		seen := map[string]struct{}{}
		for _, p := range pend {
			key := fmt.Sprintf("%s:%d", p.EventID.String(), p.AttendeeRefID)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			wl = append(wl, entity.EventWhitelist{
				EventID:       p.EventID,
				AttendeeRefID: p.AttendeeRefID,
			})
		}

		if len(wl) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
				DoNothing: true,
			}).Create(&wl).Error; err != nil {
				return err
			}
		}

		return tx.Where("attendee_ref_id = ?", refID).Delete(&entity.EventWhitelistPending{}).Error
	})
}
