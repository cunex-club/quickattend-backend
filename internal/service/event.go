package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type EventService interface {
	GetEventsService(uint64, map[string]string, context.Context) (*[]dtoRes.GetEventsIndividualEvent, *response.Pagination, *response.APIError)
}

func (s *service) GetEventsService(refID uint64, queryParams map[string]string, ctx context.Context) (*[]dtoRes.GetEventsIndividualEvent, *response.Pagination, *response.APIError) {
	pageQuery, ok := queryParams["page"]
	if !ok {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Missing required URL query parameter: page",
			Status:  400,
		}
	}
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "URL query parameter 'page' must be int",
			Status:  400,
		}
	}
	if page < 0 {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "URL query parameter 'page' must be greater than 0",
			Status:  400,
		}
	}

	size := 8
	sizeQuery, ok := queryParams["pageSize"]
	if ok {
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
	searchQuery, ok := queryParams["search"]
	if ok {
		search = strings.TrimSpace(searchQuery)
	}

	managedQuery, ok := queryParams["managed"]
	if !ok {
		res, total, hasNext, err := s.repo.Event.GetDiscoveryEvents(refID, page, size, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Uint64("ref_id", refID).
				Str("function", "EventRepository.GetDiscoveryEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting discovery events",
				Status:  500,
			}
		}
		return res, &response.Pagination{
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
	if managed {
		res, err := s.repo.Event.GetManagedEvents(refID, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Uint64("ref_id", refID).
				Str("function", "EventRepository.GetManagedEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting managed events",
				Status:  500,
			}
		}
		return res, nil, nil
	}
	res, total, hasNext, err := s.repo.Event.GetAttendedEvents(refID, page, size, search, ctx)
	if err != nil {
		s.logger.Error().Err(err).
			Uint64("ref_id", refID).
			Str("function", "EventRepository.GetAttendedEvents").
			Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
		return nil, nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error on getting attended events",
			Status:  500,
		}
	}
	return res, &response.Pagination{
		Page:     page,
		PageSize: size,
		Total:    total,
		HasNext:  hasNext,
	}, nil
}
