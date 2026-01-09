package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type EventService interface {
	GetEventsService(string, map[string]string, context.Context) (*[]dtoRes.GetEventsRes, *response.Pagination, *response.APIError)
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
	}

	formattedRes := []dtoRes.GetEventsRes{}

	managedQuery, managedOk := queryParams["managed"]
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

	managed, err := strconv.ParseBool(managedQuery)
	if err != nil {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "URL query parameter 'managed' must be boolean",
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
				Date:           (*rawResult)[i].Date,
				StartTime:      (*rawResult)[i].StartTime,
				EndTime:        (*rawResult)[i].EndTime,
				Location:       (*rawResult)[i].Location,
				Role:           (*rawResult)[i].Role,
				EvaluationForm: (*rawResult)[i].EvaluationForm,
			})
		}
	}
}
