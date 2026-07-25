package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type AuthService interface {
	GetUserService(string, context.Context) (*dtoRes.GetAuthUserRes, *response.APIError)
	VerifyCUNEXToken(string, context.Context) (*dtoRes.VerifyTokenRes, *response.APIError)
	CreateUserIfNotExists(*entity.User, context.Context) (*entity.User, *response.APIError)
}

const sessionLifetime = 8 * time.Hour

func (s *service) GetUserService(userIDStr string, ctx context.Context) (*dtoRes.GetAuthUserRes, *response.APIError) {
	uuidValidateErr := uuid.Validate(userIDStr)
	if uuidValidateErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Failed to validate user_id as UUID",
			Status:  500,
		}
	}

	userID := datatypes.UUID(datatypes.BinUUIDFromString(userIDStr))

	user, err := s.repo.Auth.GetUserById(userID, ctx)
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

	userDTO := dtoRes.GetAuthUserRes{
		ID:            user.ID.String(),
		RefID:         s.FormatRefIdToStr(user.RefID),
		UserType:      string(user.UserType),
		FirstnameTH:   user.FirstnameTH,
		SurnameTH:     user.SurnameTH,
		TitleTH:       user.TitleTH,
		FacultyNameTH: user.FacultyNameTH,
		FirstnameEN:   user.FirstnameEN,
		SurnameEN:     user.SurnameEN,
		TitleEN:       user.TitleEN,
		FacultyNameEN: user.FacultyNameEN,
	}

	return &userDTO, nil
}

func (s *service) CreateUserIfNotExists(user *entity.User, ctx context.Context) (*entity.User, *response.APIError) {
	existing, err := s.repo.Auth.GetUserByRefId(user.RefID, ctx)
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error().Err(err).Uint64("user_ref_id", user.RefID).Str("action", "query_user").Msg("query_user failed")
		return nil, &response.APIError{Code: response.ErrInternalError, Message: "internal db error", Status: 500}
	}

	created, createErr := s.repo.Auth.CreateUser(user, ctx)
	if createErr != nil {
		if errors.Is(createErr, gorm.ErrDuplicatedKey) {
			ex, _ := s.repo.Auth.GetUserByRefId(user.RefID, ctx)
			return &ex, nil
		}
		s.logger.Error().Err(createErr).Uint64("user_ref_id", user.RefID).Str("action", "create_user").Msg("create_user failed")
		return nil, &response.APIError{Code: response.ErrInternalError, Message: "failed to create user", Status: 500}
	}

	return created, nil
}

func (s *service) VerifyCUNEXToken(token string, ctx context.Context) (*dtoRes.VerifyTokenRes, *response.APIError) {
	if strings.TrimSpace(token) == "" {
		return nil, &response.APIError{
			Code:    "TOKEN_REQUIRED",
			Message: "token is required",
			Status:  400,
		}
	}

	tokenValidationUrl := "https://culab-svc.azurewebsites.net/Service.svc/profile"

	req, err := http.NewRequest("GET", tokenValidationUrl, nil)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to create token validation request",
			Status:  500,
		}
	}

	ClientId := s.cfg.LLEConfig.ProfileClientID
	if ClientId == "" {
		return nil, &response.APIError{
			Code:    "ClientId_NOT_FOUND",
			Message: "ClientId not configured",
			Status:  500,
		}
	}

	ClientSecret := s.cfg.LLEConfig.ProfileClientSecret
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

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to call external token validation API",
			Status:  500,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, &response.APIError{
			Code:    response.ErrUnauthorized,
			Message: "invalid token",
			Status:  http.StatusUnauthorized,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "unexpected response from CU NEX profile API",
			Status:  http.StatusBadGateway,
		}
	}

	var UserData entity.CUNEXProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&UserData); err != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to decode external API response",
			Status:  500,
		}
	}

	if UserData.RefId == nil || strings.TrimSpace(*UserData.RefId) == "" {
		return nil, &response.APIError{
			Code:    response.ErrUnauthorized,
			Message: "CU NEX profile does not contain a refId",
			Status:  http.StatusUnauthorized,
		}
	}
	if !isAllowedRefID(*UserData.RefId, s.cfg.BackofficeAllowedRefIDs) {
		return nil, &response.APIError{
			Code:    response.ErrForbidden,
			Message: "user is not allowed to access this backoffice",
			Status:  http.StatusForbidden,
		}
	}

	convRefId, convRefIdErr := strconv.ParseUint(*UserData.RefId, 10, 64)

	if convRefIdErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Could not convert ref_id from string to uint64",
			Status:  500,
		}
	}

	userType, validUserType := entity.ParseUserType(UserData.UserType)
	if !validUserType {
		return nil, &response.APIError{
			Code:    response.ErrUnauthorized,
			Message: "CU NEX profile returned an unsupported userType",
			Status:  http.StatusUnauthorized,
		}
	}

	User := entity.User{
		RefID:         convRefId,
		UserType:      userType,
		FirstnameTH:   UserData.FirstNameTH,
		SurnameTH:     UserData.LastNameTH,
		FirstnameEN:   UserData.FirstNameEN,
		SurnameEN:     UserData.LastNameEN,
		TitleTH:       UserData.TitleNameTH,
		TitleEN:       UserData.TitleNameEN,
		FacultyNameTH: UserData.FacultyNameTH,
		FacultyNameEN: UserData.FacultyNameEN,
	}

	// // ### MOCK USER DATA ###
	// User := entity.User{
	// 	RefID:         987654321,
	// 	FirstnameTH:   "AB",
	// 	SurnameTH:     "CD",
	// 	TitleTH:       "EEEE",
	// 	FirstnameEN:   "FG",
	// 	SurnameEN:     "HI",
	// 	TitleEN:       "JJJJ",
	// 	FacultyNameTH: "KK",
	// 	FacultyNameEN: "LL",
	// }

	upsertUser, upsertErr := s.repo.Auth.UpsertUserByRefId(&User, nil, ctx)
	if upsertErr != nil {
		s.logger.Error().
			Err(upsertErr).
			Uint64("user_ref_id", convRefId).
			Str("action", "upsert_user_by_ref_id").
			Msg("failed to upsert user by ref id")

		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Failed to upsert user by ref id",
			Status:  500,
		}
	}

	if err := s.repo.Auth.SyncWhitelistPendingToWhitelist(ctx, upsertUser.RefID); err != nil {
		s.logger.Error().
			Err(err).
			Uint64("user_ref_id", upsertUser.RefID).
			Str("action", "sync_whitelist_pending").
			Msg("failed to sync whitelist pending to whitelist")
	}

	if err := s.repo.Auth.SyncEventUserPendingToEventUser(ctx, upsertUser.ID, upsertUser.RefID); err != nil {
		s.logger.Error().
			Err(err).
			Uint64("user_ref_id", upsertUser.RefID).
			Str("action", "sync_event_user_pending").
			Msg("failed to sync event user pending to event user")
	}

	var (
		key []byte
		t   *jwt.Token
	)

	now := time.Now()
	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id":   upsertUser.ID.String(),
			"ref_id":    *UserData.RefId,
			"user_type": string(userType),
			"iat":       now.Unix(),
			"exp":       now.Add(sessionLifetime).Unix(),
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

func isAllowedRefID(refID string, allowedCSV string) bool {
	refID = strings.TrimSpace(refID)
	for _, allowed := range strings.Split(allowedCSV, ",") {
		if refID != "" && refID == strings.TrimSpace(allowed) {
			return true
		}
	}
	return false
}

func (s *service) FormatRefIdToStr(refId uint64) string {
	str := fmt.Sprint(refId)
	if len(str) < 8 {
		// In case it's staff ref id with zeros at the start
		// Student ref id cannot start with zero so it's fine
		str = strings.Repeat("0", 8-len(str)) + str
	}
	return str
}
