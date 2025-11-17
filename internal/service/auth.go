package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService interface {
	GetUserService(string) (*entity.User, *response.APIError)
	ValidateCUNEXToken(string) (*entity.CUNEXUserResponse, *response.APIError)
	CreateUserIfNotExists(*entity.User) (*entity.User, *response.APIError)
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

	ClientId, ClientIdExists := os.LookupEnv("ClientId")
	if !ClientIdExists {
		return nil, &response.APIError{
			Code:    "ClientId_NOT_FOUND",
			Message: "ClientId not configured",
			Status:  500,
		}
	}

	ClientSecret, ClientSecretExists := os.LookupEnv("ClientSecret")
	if !ClientSecretExists {
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

	if resp.StatusCode != http.StatusOK {
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

	// data := entity.CUNEXUserResponse{
	// 	UserId: "9999999",
	// 	UserType: "student",
	// 	RefId: "12345",
	// 	FirstnameEN: "Ratanon",
	// 	LastNameEN: "Khamrong",
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
