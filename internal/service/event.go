package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	b64 "encoding/base64"
	"encoding/json"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type EventService interface {
	GetParticipantService(code string, eventId string, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError)
}

func (s *service) GetParticipantService(code string, eventId string, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError) {
	if code == "" {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Missing URL path parameter 'qrcode'",
			Status:  400,
		}
	}
	if len(code) != 10 {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "URL path parameter 'qrcode' must have length of 10",
			Status:  400,
		}
	}
	numbers := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for _, r := range code {
		isDigit := false
		for _, num := range numbers {
			if string(r) == num {
				isDigit = true
			}
		}
		if !isDigit {
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL path parameter 'qrcode' contains non-number character(s)",
				Status:  400,
			}
		}
	}

	validateUuidErr := uuid.Validate(eventId)
	if validateUuidErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Invalid event ID format",
			Status:  400,
		}
	}
	eventIdUuid := datatypes.UUID(datatypes.BinUUIDFromString(eventId))

	CUNEXGetQRURL := ""
	clientId := s.cfg.LLEConfig.ClientId
	if clientId == "" {
		s.logger.Error().Str("Error", "Missing env config 'LLEClientId'")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Missing env config 'LLEClientId'",
			Status:  500,
		}
	}
	clientSecret := s.cfg.LLEConfig.ClientSecret
	if clientSecret == "" {
		s.logger.Error().Str("Error", "Missing env config 'LLEClientSecret'")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Missing env config 'LLEClientSecret'",
			Status:  500,
		}
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, formNewReqErr := http.NewRequest(http.MethodGet, CUNEXGetQRURL, nil)
	if formNewReqErr != nil {
		s.logger.Error().Err(formNewReqErr).Str("Error", "Failed to form new HTTP request for CU NEX GET qrcode")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Failed to form new HTTP request for CU NEX GET qrcode",
			Status:  500,
		}
	}

	query := req.URL.Query()
	query.Add("qrcode", code)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ClientId", clientId)
	req.Header.Set("ClientSecret", clientSecret)

	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Failed to send request for CU NEX GET qrcode",
			Status:  500,
		}
	}
	defer resp.Body.Close()

	var CUNEXSuccess entity.CUNEXGetQRSuccessResponse

	if resp.StatusCode != 200 {
		var CUNEXErr entity.CUNEXGetQRErrorResponse
		parseErr := json.NewDecoder(resp.Body).Decode(&CUNEXErr)
		if parseErr != nil {
			s.logger.Error().Err(parseErr).Str("Error", "Could not parse error response from CU NEX GET qrcode")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Could not parse error response from CU NEX GET qrcode",
				Status:  500,
			}
		}

		switch resp.StatusCode {
		case 417:
			switch CUNEXErr.ErrorCode {
			case "E001", "E002", "E003":
				// token invalid, token expired, or token inactive
				return nil, &response.APIError{
					Code:    CUNEXErr.ErrorCode,
					Message: CUNEXErr.Message,
					Status:  401,
				}
			case "E004":
				// qrcode inactive
				return nil, &response.APIError{
					Code:    CUNEXErr.ErrorCode,
					Message: CUNEXErr.Message,
					Status:  410,
				}
			case "E999":
				// internal service error
				return nil, &response.APIError{
					Code:    CUNEXErr.ErrorCode,
					Message: CUNEXErr.Message,
					Status:  500,
				}
			}

		case 401:
			s.logger.Error().Str("Error", fmt.Sprintf("Authorization error from CU NEX GET qrcode: %s", CUNEXErr.Message))
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Authorization error from CU NEX GET qrcode",
				Status:  500,
			}
		case 400:
			s.logger.Error().Str("Error", fmt.Sprintf("Bad request error from CU NEX GET qrcode: %s", CUNEXErr.Message))
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Bad request error from CU NEX GET qrcode",
				Status:  500,
			}
		case 500:
			s.logger.Error().Str("Error", fmt.Sprintf("Server error from CU NEX GET qrcode: %s", CUNEXErr.Message))
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Server error from CU NEX GET qrcode",
				Status:  500,
			}
		}
	} else {
		parseErr := json.NewDecoder(resp.Body).Decode(&CUNEXSuccess)
		if parseErr != nil {
			s.logger.Error().Err(parseErr).Str("Error", "Could not parse success response from CU NEX GET qrcode")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Could not parse success response from CU NEX GET qrcode",
				Status:  500,
			}
		}
	}

	refIdUInt, convertErr := strconv.ParseUint(CUNEXSuccess.RefId, 10, 64)
	if convertErr != nil {
		s.logger.Error().Err(convertErr).Str("Error", fmt.Sprintf("Invalid refID returned from CU NEX GET qrcode; could not convert %s to uint64", CUNEXSuccess.RefId))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid refID returned from CU NEX GET qrcode",
			Status:  500,
		}
	}

	var orgCode int64
	switch CUNEXSuccess.UserType {
	case entity.STAFFS:
		// TODO: how to get org code from staff id?
		orgCode = 0
	case entity.STUDENTS:
		code, _ := strconv.ParseInt(CUNEXSuccess.RefId[len(CUNEXSuccess.RefId)-2:], 10, 64)
		orgCode = code
	default:
		s.logger.Error().Str("Error", fmt.Sprintf("Invalid userType returned from CU NEX GET qrcode: %s", CUNEXSuccess.UserType))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid userType returned from CU NEX GET qrcode",
			Status:  500,
		}
	}

	user, attendanceType, err := s.repo.Event.GetParticipantUserAndEventInfo(ctx, eventIdUuid, refIdUInt)
	if err != nil {
		s.logger.Error().Err(err).
			Uint64("ref_id", refIdUInt).
			Str("function", "EventRepository.GetParticipantUserInfoAndAttendanceType")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	checkInTime := time.Now().UTC()

	var status string
	if attendanceType == string(entity.ALL) {
		found, err := s.repo.Event.GetParticipantCheckParticipation(ctx, eventIdUuid, refIdUInt)
		if err != nil {
			s.logger.Error().Err(err).
				Uint64("ref_id", refIdUInt).
				Str("function", "EventRepository.GetParticipantCheckParticipation")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}

		if !found {
			status = string(dtoRes.SUCCESS)
		} else {
			status = string(dtoRes.DUPLICATE)
		}

	} else if attendanceType == string(entity.WHITELIST) || attendanceType == string(entity.FACULTIES) {
		found, err := s.repo.Event.GetParticipantCheckParticipation(ctx, eventIdUuid, refIdUInt)
		if err != nil {
			s.logger.Error().Err(err).
				Uint64("ref_id", refIdUInt).
				Str("function", "EventRepository.GetParticipantCheckParticipation")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}

		if found {
			status = string(dtoRes.DUPLICATE)
		} else {
			allow, err := s.repo.Event.GetParticipantCheckAccess(ctx, orgCode, refIdUInt, attendanceType, eventIdUuid)
			if err != nil {
				s.logger.Error().Err(err).
					Uint64("ref_id", refIdUInt).
					Str("function", "EventRepository.GetParticipantCheckAccess")
				return nil, &response.APIError{
					Code:    response.ErrInternalError,
					Message: "Internal DB error",
					Status:  500,
				}
			}

			if !allow {
				status = string(dtoRes.FAIL)
			} else {
				status = string(dtoRes.SUCCESS)
			}
		}

	} else {
		s.logger.Error().
			Uint64("ref_id", refIdUInt).
			Str("Error", fmt.Sprintf("'%s' is not a possible value for attendance_type enum", attendanceType))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	rawCode := fmt.Appendf(nil, "%s.%s", checkInTime.Format("2006-01-02T15:04:05Z"), CUNEXSuccess.RefId)
	responseBody := dtoRes.GetParticipantRes{
		FirstnameTH:  user.FirstnameTH,
		SurnameTH:    user.SurnameTH,
		TitleTH:      user.TitleTH,
		FirstnameEN:  CUNEXSuccess.FirstName,
		SurnameEN:    CUNEXSuccess.LastName,
		TitleEN:      user.TitleEN,
		RefID:        refIdUInt,
		Organization: CUNEXSuccess.Organization,
		CheckInTime:  checkInTime,
		Status:       status,
		Code:         b64.StdEncoding.EncodeToString(rawCode),
	}

	return &responseBody, nil
}
