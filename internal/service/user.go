package service

import (
	"context"
	"errors"
	"strconv"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserByRefId(refIdStr string, ctx context.Context) (*dtoRes.GetUserByRefIdRes, *response.APIError)
}

func (s *service) GetUserByRefId(refIdStr string, ctx context.Context) (*dtoRes.GetUserByRefIdRes, *response.APIError) {
	refId, err := strconv.ParseUint(refIdStr, 10, 64)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Status:  400,
			Message: "Invalid ref_id format",
		}
	}

	user, err := s.repo.Auth.GetUserByRefId(refId, ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Status:  404,
				Message: "User not found",
			}
		}

		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Status:  500,
			Message: "Internal DB error",
		}
	}

	return &dtoRes.GetUserByRefIdRes{
		RefID:           s.FormatRefIdToStr(user.RefID),
		FirstnameTH:     user.FirstnameTH,
		SurnameTH:       user.SurnameTH,
		TitleTH:         user.TitleTH,
		FacultyNameTH:   user.FacultyNameTH,
		FirstnameEN:     user.FirstnameEN,
		SurnameEN:       user.SurnameEN,
		TitleEN:         user.TitleEN,
		FacultyNameEN:   user.FacultyNameEN,
		ProfileImageURL: user.ProfileImageURL,
	}, nil
}
