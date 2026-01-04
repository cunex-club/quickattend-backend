package service

import (
	"context"
	"errors"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
)

type AuthService interface {
	GetUserService(uint64, context.Context) (*entity.User, *response.APIError)
	VerifyCUNEXToken(string, context.Context) (*dtoRes.VerifyTokenRes, *response.APIError)
	CreateUserIfNotExists(*entity.User, context.Context) (*entity.User, *response.APIError)
}

func (s *service) GetUserService(refID uint64, ctx context.Context) (*entity.User, *response.APIError) {
	user, err := s.repo.Auth.GetUserByRefId(refID, ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &response.APIError{
			Code:    response.ErrNotFound,
			Message: "User not found",
			Status:  404,
		}
	}
	if err != nil {
		s.logger.Error().Err(err).Msg("Internal DB error")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	return &user, nil
}

func (s *service) CreateUserIfNotExists(user *entity.User, ctx context.Context) (*entity.User, *response.APIError) {
	foundUser, findErr := s.repo.Auth.GetUserByRefId(user.RefID, ctx)
	if findErr == nil {
		return &foundUser, nil
	}

	if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		s.logger.Error().
			Err(findErr).
			Uint64("user_ref_id", user.RefID).
			Str("action", "query_user").
			Msg("service failed to query user by ref_id")

		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	createdUser, createErr := s.repo.Auth.CreateUser(user, ctx)
	if createErr != nil {
		if errors.Is(createErr, gorm.ErrDuplicatedKey) {
			existingUser, _ := s.repo.Auth.GetUserByRefId(user.RefID, ctx)
			return &existingUser, nil
		}

		s.logger.Error().
			Err(createErr).
			Uint64("user_ref_id", user.RefID).
			Str("action", "create_user").
			Msg("service failed to create user")

		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to create user",
			Status:  500,
		}
	}

	return createdUser, nil
}

func (s *service) VerifyCUNEXToken(token string, ctx context.Context) (*dtoRes.VerifyTokenRes, *response.APIError) {
	// if strings.TrimSpace(token) == "" {
	// 	return nil, &response.APIError{
	// 		Code:    "TOKEN_REQUIRED",
	// 		Message: "token is required",
	// 		Status:  400,
	// 	}
	// }

	// tokenValidationUrl := ""

	// client := &http.Client{}
	// req, err := http.NewRequest("GET", tokenValidationUrl, nil)
	// if err != nil {
	// 	return nil, &response.APIError{
	// 		Code:    response.ErrInternalError,
	// 		Message: "failed to create token validation request",
	// 		Status:  500,
	// 	}
	// }

	// ClientId := s.cfg.LLEConfig.ClientId
	// if ClientId == "" {
	// 	return nil, &response.APIError{
	// 		Code:    "ClientId_NOT_FOUND",
	// 		Message: "ClientId not configured",
	// 		Status:  500,
	// 	}
	// }

	// ClientSecret := s.cfg.LLEConfig.ClientSecret
	// if ClientSecret == "" {
	// 	return nil, &response.APIError{
	// 		Code:    "ClientSecret_NOT_FOUND",
	// 		Message: "ClientSecret not configured",
	// 		Status:  500,
	// 	}
	// }

	// req.Header.Set("Content-type", "application/json")
	// req.Header.Set("ClientId", ClientId)
	// req.Header.Set("ClientSecret", ClientSecret)

	// q := req.URL.Query()
	// q.Add("token", token)
	// req.URL.RawQuery = q.Encode()

	// resp, err := client.Do(req)
	// if err != nil {
	// 	return nil, &response.APIError{
	// 		Code:    response.ErrInternalError,
	// 		Message: "failed to call external token validation API",
	// 		Status:  500,
	// 	}
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode == http.StatusExpectationFailed {
	// 	return nil, &response.APIError{
	// 		Code:    response.ErrUnauthorized,
	// 		Message: "invalid token",
	// 		Status:  resp.StatusCode,
	// 	}
	// }

	// var UserData entity.CUNEXUserResponse
	// if err := json.NewDecoder(resp.Body).Decode(&UserData); err != nil {
	// 	return nil, &response.APIError{
	// 		Code:    response.ErrInternalError,
	// 		Message: "failed to decode external API response",
	// 		Status:  500,
	// 	}
	// }

	// convRefId, convRefIdErr := strconv.ParseUint(UserData.RefId, 10, 64)

	// if convRefIdErr != nil {
	// 	return nil, &response.APIError{
	// 		Code:    response.ErrInternalError,
	// 		Message: "Could not convert ref_id from string to uint64",
	// 		Status:  500,
	// 	}
	// }

	// User := entity.User{
	// 	RefID:       convRefId,
	// 	FirstnameTH: UserData.FirstNameTH,
	// 	SurnameTH:   UserData.LastNameTH,
	// 	TitleTH:     "",
	// 	FirstnameEN: UserData.FirstnameEN,
	// 	SurnameEN:   UserData.LastNameEN,
	// 	TitleEN:     "",
	// }

	// ### MOCK USER DATA ###
	// User := entity.User{
	// 	RefID:       987654321,
	// 	FirstnameTH: "AB",
	// 	SurnameTH:   "CD",
	// 	TitleTH:     "EEEE",
	// 	FirstnameEN: "FG",
	// 	SurnameEN:   "HI",
	// 	TitleEN:     "JJJJ",
	// }
	User := entity.User{
		RefID:       6874440950,
		FirstnameTH: "ค",
		SurnameTH:   "ง",
		TitleTH:     "นางสาว",
		FirstnameEN: "C",
		SurnameEN:   "D",
		TitleEN:     "MS",
	}

	createdUser, createdUserErr := s.CreateUserIfNotExists(&User, ctx)
	if createdUserErr != nil {
		return nil, &response.APIError{
			Code:    createdUserErr.Code,
			Message: createdUserErr.Message,
			Status:  createdUserErr.Status,
		}
	}

	var (
		key []byte
		t   *jwt.Token
	)

	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"ref_id": createdUser.RefID,
		})

	JWTSecret := s.cfg.JWTSecret
	if JWTSecret == "" {
		return nil, &response.APIError{
			Code:    "JWT_SIGN_KEY_NOT_FOUND",
			Message: "JWT signing key not configured",
			Status:  500,
		}
	}

	key = []byte(JWTSecret)
	access_token, signErr := t.SignedString(key)
	if signErr != nil {
		return nil, &response.APIError{
			Code:    "JWT_SIGN_FAIL",
			Message: "failed to sign token",
			Status:  500,
		}
	}

	return &dtoRes.VerifyTokenRes{
		AccessToken: access_token,
	}, nil
}
