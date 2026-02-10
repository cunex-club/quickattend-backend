package repository

import (
	"context"
	"fmt"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, payload entity.CreateEventPayload) (*dtoRes.CreateEventRes, error)
	UpdateEvent(ctx context.Context, id string, payload entity.CreateEventPayload) (*dtoRes.UpdateEventRes, error)
}

func (r *repository) CreateEvent(ctx context.Context, payload entity.CreateEventPayload) (*dtoRes.CreateEventRes, error) {
	var res dtoRes.CreateEventRes

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&payload.Event).Error; err != nil {
			return err
		}

		if len(payload.Agendas) > 0 {
			for i := range payload.Agendas {
				payload.Agendas[i].EventID = payload.Event.ID
			}
			if err := tx.Create(&payload.Agendas).Error; err != nil {
				return err
			}
		}

		if len(payload.Whitelist) > 0 {
			for i := range payload.Whitelist {
				payload.Whitelist[i].EventID = payload.Event.ID
			}

			wlOK, wlPend, err := splitWhitelistAndPending(ctx, tx, payload.Whitelist)
			if err != nil {
				return err
			}

			if len(wlOK) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlOK).Error; err != nil {
					return err
				}
			}

			if len(wlPend) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlPend).Error; err != nil {
					return err
				}
			}
		}

		if len(payload.AllowedFaculties) > 0 {
			for i := range payload.AllowedFaculties {
				payload.AllowedFaculties[i].EventID = payload.Event.ID
			}
			if err := tx.Create(&payload.AllowedFaculties).Error; err != nil {
				return err
			}
		}

		eventUsers, err := buildEventUsersFromInput(ctx, tx, payload.Event.ID, payload.EventUsersInput)
		if err != nil {
			return err
		}
		if len(eventUsers) > 0 {
			if err := tx.Create(&eventUsers).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	res = dtoRes.CreateEventRes{
		ID: payload.Event.ID.String(),
	}
	return &res, nil
}

func (r *repository) UpdateEvent(ctx context.Context, id string, payload entity.CreateEventPayload) (*dtoRes.UpdateEventRes, error) {
	var res dtoRes.UpdateEventRes

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.Event
		if err := tx.First(&existing, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Model(&existing).Updates(map[string]any{
			"name":              payload.Event.Name,
			"organizer":         payload.Event.Organizer,
			"description":       payload.Event.Description,
			"date":              payload.Event.Date,
			"start_time":        payload.Event.StartTime,
			"end_time":          payload.Event.EndTime,
			"location":          payload.Event.Location,
			"attendance_type":   payload.Event.AttendanceType,
			"allow_all_to_scan": payload.Event.AllowAllToScan,
			"evaluation_form":   payload.Event.EvaluationForm,
			"revealed_fields":   payload.Event.RevealedFields,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventAgenda{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventWhitelist{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventWhitelistPending{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventAllowedFaculties{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventUser{}).Error; err != nil {
			return err
		}

		if len(payload.Agendas) > 0 {
			for i := range payload.Agendas {
				payload.Agendas[i].EventID = existing.ID
			}
			if err := tx.Create(&payload.Agendas).Error; err != nil {
				return err
			}
		}

		if len(payload.Whitelist) > 0 {
			for i := range payload.Whitelist {
				payload.Whitelist[i].EventID = existing.ID
			}

			wlOK, wlPend, err := splitWhitelistAndPending(ctx, tx, payload.Whitelist)
			if err != nil {
				return err
			}

			if len(wlOK) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlOK).Error; err != nil {
					return err
				}
			}

			if len(wlPend) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlPend).Error; err != nil {
					return err
				}
			}
		}

		if len(payload.AllowedFaculties) > 0 {
			for i := range payload.AllowedFaculties {
				payload.AllowedFaculties[i].EventID = existing.ID
			}
			if err := tx.Create(&payload.AllowedFaculties).Error; err != nil {
				return err
			}
		}

		eventUsers, err := buildEventUsersFromInput(ctx, tx, existing.ID, payload.EventUsersInput)
		if err != nil {
			return err
		}
		if len(eventUsers) > 0 {
			if err := tx.Create(&eventUsers).Error; err != nil {
				return err
			}
		}

		res.ID = existing.ID.String()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &res, nil
}

func splitWhitelistAndPending(ctx context.Context, tx *gorm.DB, wl []entity.EventWhitelist) ([]entity.EventWhitelist, []entity.EventWhitelistPending, error) {
	if len(wl) == 0 {
		return nil, nil, nil
	}

	refIDs := make([]uint64, 0, len(wl))
	seenRef := map[uint64]struct{}{}
	for _, x := range wl {
		if _, ok := seenRef[x.AttendeeRefID]; ok {
			continue
		}
		seenRef[x.AttendeeRefID] = struct{}{}
		refIDs = append(refIDs, x.AttendeeRefID)
	}

	var existing []uint64
	if err := tx.WithContext(ctx).
		Model(&entity.User{}).
		Where("ref_id IN ?", refIDs).
		Pluck("ref_id", &existing).Error; err != nil {
		return nil, nil, err
	}

	existSet := map[uint64]struct{}{}
	for _, id := range existing {
		existSet[id] = struct{}{}
	}

	okOut := make([]entity.EventWhitelist, 0, len(wl))
	pendOut := make([]entity.EventWhitelistPending, 0)

	seenPairOK := map[string]struct{}{}
	seenPairPend := map[string]struct{}{}

	for _, x := range wl {
		key := fmt.Sprintf("%s:%d", x.EventID.String(), x.AttendeeRefID)

		if _, ok := existSet[x.AttendeeRefID]; ok {
			if _, dup := seenPairOK[key]; dup {
				continue
			}
			seenPairOK[key] = struct{}{}
			okOut = append(okOut, x)
			continue
		}

		if _, dup := seenPairPend[key]; dup {
			continue
		}
		seenPairPend[key] = struct{}{}
		pendOut = append(pendOut, entity.EventWhitelistPending{
			EventID:       x.EventID,
			AttendeeRefID: x.AttendeeRefID,
		})
	}

	return okOut, pendOut, nil
}

// buildEventUsersFromInput maps ref_id -> user_id(uuid) แล้วสร้าง []EventUser
func buildEventUsersFromInput(ctx context.Context, tx *gorm.DB, eventID datatypes.UUID, in []entity.EventUserInput) ([]entity.EventUser, error) {
	if len(in) == 0 {
		return nil, nil
	}

	refIDs := make([]uint64, 0, len(in))
	seen := map[uint64]struct{}{}
	for _, x := range in {
		if _, ok := seen[x.RefID]; ok {
			continue
		}
		seen[x.RefID] = struct{}{}
		refIDs = append(refIDs, x.RefID)
	}

	var users []entity.User
	if err := tx.WithContext(ctx).Where("ref_id IN ?", refIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	userByRef := make(map[uint64]entity.User, len(users))
	for _, u := range users {
		userByRef[u.RefID] = u
	}

	out := make([]entity.EventUser, 0, len(in))
	seenPair := map[string]struct{}{}
	for _, x := range in {
		u, ok := userByRef[x.RefID]
		if !ok {
			return nil, fmt.Errorf("unknown ref_id in managers_and_staff: %d", x.RefID)
		}

		key := fmt.Sprintf("%d:%s", x.RefID, x.Role)
		if _, ok := seenPair[key]; ok {
			continue
		}
		seenPair[key] = struct{}{}

		out = append(out, entity.EventUser{
			EventID: eventID,
			UserID:  u.ID,
			Role:    x.Role,
		})
	}

	return out, nil
}
