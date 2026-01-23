package service

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type EventService interface {
	DeleteById(EventID string, ctx context.Context) *response.APIError
	DuplicateById(EventID string, ctx context.Context) (*entity.Event, *response.APIError)
	CheckIn(checkInReq *dtoReq.CheckInReq, ctx context.Context) *response.APIError
}

func (s *service) CheckIn(checkInReq *dtoReq.CheckInReq, ctx context.Context) *response.APIError {

	decoded, err := base64.StdEncoding.DecodeString(checkInReq.EncodedOneTimeCode)
	if err != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "failed to interpret one_time_code as base64 encoded",
			Status:  400,
		}
	}

	raw := string(decoded)
	idx := strings.LastIndex(raw, ".")
	if idx == -1 {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "invalid one_time_code format",
			Status:  400,
		}
	}

	strTimeStamp := raw[:idx]
	strCheckInRowId := raw[idx+1:]

	checkInRowId, err := uuid.Parse(strCheckInRowId)
	if err != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "failed to convert checkInRowId to uuid",
			Status:  400,
		}
	}

	timeStamp, err := time.Parse(time.RFC3339, strTimeStamp)
	if err != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "failed to convert timeStamp to time (go)",
			Status:  400,
		}
	}

	s.logger.Info().
		Str("timeStamp", timeStamp.String()).
		Str("checkInRowId", checkInRowId.String()).
		Msg("Received timeStamp and target row-id to check-in Event-Participant")

	// Check-in
	if err := s.repo.Event.CheckIn(
		checkInRowId,
		timeStamp,
		checkInReq.Comment,
		ctx,
	); err != nil {

		if errors.Is(err, entity.ErrAlreadyCheckedIn) || errors.Is(err, entity.ErrCheckInTargetNotFound) {
			return &response.APIError{
				Code:    response.ErrBadRequest,
				Message: err.Error(),
				Status:  400,
			}
		}

		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	return nil
}

func (s *service) DeleteById(EventId string, ctx context.Context) *response.APIError {
	event_id, parseErr := uuid.Parse(EventId)
	if parseErr != nil {
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to parse event_id to uuid",
			Status:  500,
		}
	}

	eventDeleteErr := s.repo.Event.DeleteById(event_id, ctx)

	if errors.Is(eventDeleteErr, gorm.ErrRecordNotFound) {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("event not found")
		return &response.APIError{
			Code:    response.ErrNotFound,
			Message: "event not found",
			Status:  404,
		}
	}

	if errors.Is(eventDeleteErr, entity.ErrNilUUID) {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("attempt deleting nil uuid")
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "nil uuid not allowed",
			Status:  400,
		}
	}

	if eventDeleteErr != nil {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("service failed to delete event")
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	return nil
}

func (s *service) DuplicateById(EventId string, ctx context.Context) (*entity.Event, *response.APIError) {
	event_id, parseErr := uuid.Parse(EventId)
	if parseErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to parse event_id to uuid",
			Status:  400,
		}
	}

	originalEvent, findErr := s.repo.Event.FindById(event_id, ctx)
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "specified event not found",
				Status:  400,
			}
		}
		s.logger.Error().
			Err(findErr).
			Str("event_id", EventId).
			Str("action", "duplicate_event_find").
			Msg("failed to find event for duplication")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	newEvent := *originalEvent
	newEvent.ID = datatypes.UUID(uuid.New())

	// breaking the memory link from originalEvent
	// Whitelist
	newEvent.EventWhitelist = make([]entity.EventWhitelist, 0, len(originalEvent.EventWhitelist))
	for _, item := range originalEvent.EventWhitelist {
		newEvent.EventWhitelist = append(newEvent.EventWhitelist, entity.EventWhitelist{
			AttendeeRefID: item.AttendeeRefID,
		})
	}

	// Faculties
	newEvent.EventAllowedFaculties = make([]entity.EventAllowedFaculties, 0, len(originalEvent.EventAllowedFaculties))
	for _, item := range originalEvent.EventAllowedFaculties {
		newEvent.EventAllowedFaculties = append(newEvent.EventAllowedFaculties, entity.EventAllowedFaculties{
			FacultyNO: item.FacultyNO,
		})
	}

	// Agenda
	newEvent.EventAgenda = make([]entity.EventAgenda, 0, len(originalEvent.EventAgenda))
	for _, item := range originalEvent.EventAgenda {
		newEvent.EventAgenda = append(newEvent.EventAgenda, entity.EventAgenda{
			ActivityName: item.ActivityName,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
		})
	}

	createdEvent, createErr := s.repo.Event.Create(&newEvent, ctx)
	if createErr != nil {
		s.logger.Error().
			Err(createErr).
			Str("event_id", EventId).
			Str("action", "duplicate_event").
			Msg("failed to duplicate event")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to duplicate event",
			Status:  500,
		}
	}

	return createdEvent, nil
}
