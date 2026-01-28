package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type EventService interface {
	GetOneEventService(eventIdStr string, userIdStr string, ctx context.Context) (res *dtoRes.GetOneEventRes, err *response.APIError)
	GetEventsService(userIDStr string, queryParams map[string]string, ctx context.Context) (*[]dtoRes.GetEventsRes, *response.Pagination, *response.APIError)
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

func (s *service) GetEventsService(userIDStr string, queryParams map[string]string, ctx context.Context) (*[]dtoRes.GetEventsRes, *response.Pagination, *response.APIError) {
	uuidValidationErr := uuid.Validate(userIDStr)
	if uuidValidationErr != nil {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Invalid UUID format for user_id from middleware",
			Status:  500,
		}
	}
	userID := datatypes.UUID(datatypes.BinUUIDFromString(userIDStr))

	pageQuery, pageOk := queryParams["page"]
	var page int
	if pageOk {
		pageInt, err := strconv.Atoi(pageQuery)
		if err != nil {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'page' must be int",
				Status:  400,
			}
		}
		if pageInt < 0 {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'page' must be greater than 0",
				Status:  400,
			}
		}
		page = pageInt
	}

	size := 8
	sizeQuery, sizeOk := queryParams["pageSize"]
	if sizeOk {
		pageSizeInt, err := strconv.Atoi(sizeQuery)
		if err != nil {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'pageSize' must be int",
				Status:  400,
			}
		}
		if pageSizeInt < 1 || pageSizeInt > 10 {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'pageSize' must be within range [1, 10]",
				Status:  400,
			}
		}
		size = pageSizeInt
	}

	search := ""
	searchQuery, searchOk := queryParams["search"]
	if searchOk {
		search = strings.TrimSpace(searchQuery)
		if utf8.RuneCountInString(search) > 256 {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'search' longer than 256 characters",
				Status:  400,
			}
		}
	}

	formattedRes := []dtoRes.GetEventsRes{}

	managedQuery, managedOk := queryParams["managed"]
	// 'managed' not present -> get discovery events
	if !managedOk {
		if !pageOk {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "Missing required URL query parameter: page",
				Status:  400,
			}
		}
		res, total, hasNext, err := s.repo.Event.GetDiscoveryEvents(userID, page, size, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userIDStr).
				Str("function", "EventRepository.GetDiscoveryEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting discovery events",
				Status:  500,
			}
		}
		s._GetEventsDTOFormat(res, &formattedRes)
		return &formattedRes, &response.Pagination{
			Page:     page,
			PageSize: size,
			Total:    total,
			HasNext:  hasNext,
		}, nil
	}

	// 'managed' present -> parse and get managed or participated events
	managed, err := strconv.ParseBool(managedQuery)
	if err != nil {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "URL query parameter 'managed' must be boolean",
			Status:  400,
		}
	}
	switch managed {
	case true:
		res, err := s.repo.Event.GetManagedEvents(userID, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userIDStr).
				Str("function", "EventRepository.GetManagedEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting managed events",
				Status:  500,
			}
		}
		s._GetEventsDTOFormat(res, &formattedRes)
		return &formattedRes, nil, nil

	default:
		if !pageOk {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "Missing required URL query parameter: page",
				Status:  400,
			}
		}
		res, total, hasNext, err := s.repo.Event.GetAttendedEvents(userID, page, size, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userIDStr).
				Str("function", "EventRepository.GetAttendedEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting attended events",
				Status:  500,
			}
		}
		s._GetEventsDTOFormat(res, &formattedRes)
		return &formattedRes, &response.Pagination{
			Page:     page,
			PageSize: size,
			Total:    total,
			HasNext:  hasNext,
		}, nil
	}
}

func (s *service) _GetEventsDTOFormat(rawResult *[]entity.GetEventsQueryResult, result *[]dtoRes.GetEventsRes) {
	length := len(*rawResult)
	if length > 0 {
		for i := 0; i < length; i++ {
			*result = append(*result, dtoRes.GetEventsRes{
				ID:             (*rawResult)[i].ID.String(),
				Name:           (*rawResult)[i].Name,
				Organizer:      (*rawResult)[i].Organizer,
				Description:    (*rawResult)[i].Description,
				StartTime:      (*rawResult)[i].StartTime.UTC(),
				EndTime:        (*rawResult)[i].EndTime.UTC(),
				Location:       (*rawResult)[i].Location,
				Role:           (*rawResult)[i].Role,
				EvaluationForm: (*rawResult)[i].EvaluationForm,
			})
		}
	}
}
