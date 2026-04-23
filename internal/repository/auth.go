package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthRepository interface {
	GetUserById(datatypes.UUID, context.Context) (entity.User, error)
	GetUserByRefId(uint64, context.Context) (entity.User, error)
	CreateUser(*entity.User, context.Context) (*entity.User, error)

	// If user with given `ref_id` doesn't exist, create. Otherwise update all fields, excluding ones specified in `fieldsToOmit`.
	// If `fieldsToOmit` is nil, all fields of the user are updated expect ID.
	UpsertUserByRefId(user *entity.User, fieldsToOmit *[]string, ctx context.Context) (*entity.User, error)

	FindWhitelistPendingByRefID(ctx context.Context, refID uint64) ([]entity.EventWhitelistPending, error)
	DeleteWhitelistPendingByRefID(ctx context.Context, refID uint64) error
	SyncWhitelistPendingToWhitelist(ctx context.Context, refID uint64) error
}

func (r *repository) GetUserById(userID datatypes.UUID, ctx context.Context) (entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, &entity.User{ID: userID}).Error
	return user, err
}

func (r *repository) GetUserByRefId(refID uint64, ctx context.Context) (entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, &entity.User{RefID: refID}).Error
	return user, err
}

func (r *repository) CreateUser(user *entity.User, ctx context.Context) (*entity.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, err
}

func (r *repository) UpsertUserByRefId(user *entity.User, fieldsToOmit *[]string, ctx context.Context) (*entity.User, error) {
	db := r.db.WithContext(ctx)

	var exist bool
	err := db.Raw(`SELECT EXISTS 
		(SELECT 1 FROM users WHERE ref_id = ?) 
		AS subQuery`, user.RefID).
		Scan(&exist).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return user, err
	}

	if !exist {
		return user, db.Model(&entity.User{}).Create(user).Error
	}

	// Update
	userMap := map[string]any{
		"ref_id":            user.RefID,
		"firstname_th":      user.FirstnameTH,
		"surname_th":        user.SurnameTH,
		"title_th":          user.TitleTH,
		"faculty_name_th":   user.FacultyNameTH,
		"firstname_en":      user.FirstnameEN,
		"surname_en":        user.SurnameEN,
		"title_en":          user.TitleEN,
		"faculty_name_en":   user.FacultyNameEN,
		"profile_image_url": user.ProfileImageURL,
	}
	if fieldsToOmit != nil {
		for _, field := range *fieldsToOmit {
			delete(userMap, field)
		}
	}

	update := db.Model(&entity.User{}).
		Where("ref_id = ?", user.RefID).
		Updates(userMap)
	return user, update.Error
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
