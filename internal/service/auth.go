package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"gorm.io/gorm"
)

type AuthService interface {
	GetUserService(uint64) (*entity.User, *response.APIError)
	ValidateCUNEXToken(string) (*entity.CUNEXUserResponse, *response.APIError)
	CreateUserIfNotExists(*entity.User) (*entity.User, *response.APIError)
}

func (s *service) GetUserService(refID uint64) (*entity.User, *response.APIError) {
	user, err := s.repo.Auth.GetUser(refID)
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

func (s *service) ValidateCUNEXToken(token string) (*entity.CUNEXUserResponse, *response.APIError) {
	url := "https://jsonplaceholder.typicode.com/todos/1"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to create token validation request",
			Status:  500,
		}
	}

	ClientId := config.Load().LLEConfig.ClientId
	if ClientId == "" {
		return nil, &response.APIError{
			Code:    "ClientId_NOT_FOUND",
			Message: "ClientId not configured",
			Status:  500,
		}
	}

	ClientSecret := config.Load().LLEConfig.ClientSecret
	if ClientSecret == "" {
		return nil, &response.APIError{
			Code:    "ClientSecret_NOT_FOUND",
			Message: "ClientSecret not configured",
			Status:  500,
		}
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("ClientId", ClientId)
	req.Header.Set("ClientSecret", ClientSecret)

	q := req.URL.Query()
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to call external token validation API",
			Status:  500,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusExpectationFailed {
		return nil, &response.APIError{
			Code:    response.ErrUnauthorized,
			Message: "invalid token",
			Status:  resp.StatusCode,
		}
	}

	var data entity.CUNEXUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to decode external API response",
			Status:  500,
		}
	}

	// MOCK USER DATA
	// data := entity.CUNEXUserResponse{
	// 	UserId: "9999999",
	// 	UserType: "student",
	// 	RefId: "12345",
	// 	FirstnameEN: "Somchai",
	// 	LastNameEN: "Sawasdee",
	// 	FirstNameTH: "dddd",
	// 	LastNameTH: "eeee",
	// }

	return &data, nil
}

func (s *service) CreateUserIfNotExists(user *entity.User) (*entity.User, *response.APIError) {
	foundUser, err := s.repo.Auth.GetUser(user.RefID)
	if err == nil {
		return &foundUser, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	createdUser, createErr := s.repo.Auth.CreateUser(user)
	if createErr != nil {
		if errors.Is(createErr, gorm.ErrDuplicatedKey) {
			existingUser, _ := s.repo.Auth.GetUser(user.RefID)
			return &existingUser, nil
		}

		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to create user",
			Status:  500,
		}
	}

	return createdUser, nil
}
