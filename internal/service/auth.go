package service

import (
	"errors"
	"os"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService interface {
	GetUserService(string) (*entity.User, *response.APIError)
}

func (s *service) GetUserService(tokenStr string) (*entity.User, *response.APIError) {
	var secretKey, _ = os.LookupEnv("JWT_KEY")

	result, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("Unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrUnauthorized,
			Message: "Unexpected signing method",
			Status:  401,
		}
	}

	claims, ok := result.Claims.(jwt.MapClaims)
	if ok && result.Valid {
		ref_id, ok := claims["ref_id"].(uint64)

		if !ok {
			return nil, &response.APIError{
				Code:    response.ErrUnauthorized,
				Message: "ref_id not found in token",
				Status:  401,
			}
		}

		user, err := s.repo.Auth.GetUser(ref_id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Message: "User not found",
				Status:  404,
			}
		}
		if err != nil {
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}

		return &user, nil
	}

	return nil, &response.APIError{
		Code:    response.ErrUnauthorized,
		Message: "Invalid jwt token",
		Status:  401,
	}
}
