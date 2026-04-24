package repository

import (
	"context"
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

	// If user with given `ref_id` doesn't exist, create and fill in the ID field.
	// Otherwise update all fields, excluding ID and those specified in `fieldsToOmit`, and return the updated entity.
	//
	// If `fieldsToOmit` is nil, all fields except ID are updated.
	UpsertUserByRefId(user *entity.User, fieldsToOmit *[]string, ctx context.Context) (*entity.User, error)

	FindWhitelistPendingByRefID(ctx context.Context, refID uint64) ([]entity.EventWhitelistPending, error)
	DeleteWhitelistPendingByRefID(ctx context.Context, refID uint64) error
	SyncWhitelistPendingToWhitelist(ctx context.Context, refID uint64) error

	SyncEventUserPendingToEventUser(ctx context.Context, userID datatypes.UUID, refID uint64) error
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

	err := r.db.WithContext(ctx).Clauses(
		clause.Returning{},
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "ref_id"}},
			DoUpdates: clause.Assignments(userMap),
		},
	).Create(&user).Error
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

func (r *repository) SyncEventUserPendingToEventUser(ctx context.Context, userID datatypes.UUID, refID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pending []entity.EventUserPending
		if err := tx.Where("user_ref_id = ?", refID).Find(&pending).Error; err != nil {
			return err
		}
		if len(pending) == 0 {
			return nil
		}

		eu := make([]entity.EventUser, 0, len(pending))
		seen := map[string]struct{}{}
		for _, r := range pending {
			key := fmt.Sprintf("%d:%s", r.UserRefID, r.EventID.String())
			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}
			eu = append(eu, entity.EventUser{
				UserID:  userID,
				EventID: r.EventID,
				Role:    r.Role,
			})
		}

		if len(eu) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "event_id"}},
				DoNothing: true,
			}).Create(&eu).Error; err != nil {
				return err
			}
		}

		return tx.Where("user_ref_id = ?", refID).Delete(&entity.EventUserPending{}).Error
	})
}
