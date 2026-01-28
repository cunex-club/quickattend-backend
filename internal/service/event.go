package service

import (
	"context"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type EventService interface {
	GetOneEventService(eventIdStr string, userIdStr string, ctx context.Context) (res *dtoRes.GetOneEventRes, err *response.APIError)
}

func (s *service) GetOneEventService(eventIdStr string, userIdStr string, ctx context.Context) (*dtoRes.GetOneEventRes, *response.APIError) {
	eventIdErr := uuid.Validate(eventIdStr)
	if eventIdErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Invalid URL path parameter 'id'",
			Status:  400,
		}
	}
	eventId := datatypes.UUID(datatypes.BinUUIDFromString(eventIdStr))

	userIdErr := uuid.Validate(userIdStr)
	if userIdErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid user_id from JWT claim",
			Status:  500,
		}
	}
	userId := datatypes.UUID(datatypes.BinUUIDFromString(userIdStr))

	eventWithCount, agenda, err := s.repo.Event.GetOneEvent(eventId, userId, ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Message: "Event with this id not found",
				Status:  404,
			}
		} else {
			s.logger.Error().Err(err).
				Str("function", "EventRepository.GetOneEvent")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}
	}

	agendaDTO := []dtoRes.GetOneEventAgenda{}
	if len(*agenda) > 0 {
		for _, slot := range *agenda {
			agendaDTO = append(agendaDTO, dtoRes.GetOneEventAgenda{
				ActivityName: slot.ActivityName,
				StartTime:    slot.StartTime.UTC(),
				EndTime:      slot.EndTime.UTC(),
			})
		}
	}

	finalRes := dtoRes.GetOneEventRes{
		Name:            eventWithCount.Name,
		Organizer:       eventWithCount.Organizer,
		Description:     eventWithCount.Description,
		StartTime:       eventWithCount.StartTime.UTC(),
		EndTime:         eventWithCount.EndTime.UTC(),
		Location:        eventWithCount.Location,
		TotalRegistered: eventWithCount.TotalRegistered,
		EvaluationForm:  eventWithCount.EvaluationForm,
		Agenda:          agendaDTO,
		Role:            eventWithCount.Role,
	}

	return &finalRes, nil

}
