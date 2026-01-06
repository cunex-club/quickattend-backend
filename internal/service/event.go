package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	b64 "encoding/base64"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"gorm.io/datatypes"
)

type EventService interface {
	GetParticipantService(code string, eventId string, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError)
	ValidateQRCode(code string) (errMsg string)
}

func (s *service) GetParticipantService(code string, eventId string, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError) {
	if code == "" {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Missing URL path parameter 'qrcode'",
			Status:  400,
		}
	}
	length := 0
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
		length += 1
	}
	if length != 10 {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "URL path parameter 'qrcode' must have length of 10",
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
	query.Set("qrcode", code)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ClientId", clientId)
	req.Header.Set("ClientSecret", clientSecret)

	resp, _ := client.Do(req)
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
			// Need specific error message?
			return nil, &response.APIError{
				Code:    CUNEXErr.ErrorCode,
				Message: fmt.Sprintf("QR code related error from CU NEX GET qrcode: %s", CUNEXErr.Message),
				Status:  417,
			}
		// In other cases, count as "internal error" with specific details,
		// because if client provides valid resources but CU NEX server declines, the fault is at our own server
		case 408:
			s.logger.Error().Str("Error", fmt.Sprintf("Time out error from CU NEX GET qrcode: %s", CUNEXErr.Message))
			return nil, &response.APIError{
				Code:    "TIMEOUT",
				Message: "Time out error from CU NEX GET qrcode",
				Status:  500,
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
	defer resp.Body.Close()

	refIdUInt, convertErr := strconv.ParseUint(CUNEXSuccess.RefId, 10, 64)
	if convertErr != nil {
		s.logger.Error().Err(convertErr).Str("Error", "Invalid refID returned from CU NEX GET qrcode; could not convert to uint64")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid refID returned from CU NEX GET qrcode; could not convert to uint64",
			Status:  500,
		}
	}

	var orgCode int64
	switch CUNEXSuccess.UserType {
	case entity.STAFFS:
		// TODO: how to get org code from staff id?
		orgCode = 0
	case entity.STUDENTS:
		code, _ := strconv.ParseInt(CUNEXSuccess.RefId[len(CUNEXSuccess.RefId)-2:len(CUNEXSuccess.RefId)], 10, 64)
		orgCode = code
	default:
		s.logger.Error().Str("Error", fmt.Sprintf("Invalid userType returned from CU NEX GET qrcode: %s", CUNEXSuccess.UserType))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid userType returned from CU NEX GET qrcode",
			Status:  500,
		}
	}

	user, attendanceType, err := s.repo.Event.GetParticipantUserInfoAndAttendanceType(ctx, eventIdUuid, refIdUInt)
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
			allow, err := s.repo.Event.GetParticipantCheckAccess(ctx, orgCode, attendanceType, eventIdUuid)
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
			Str("Error", fmt.Sprintf("'%s' not within possible value of attendance_type enum", attendanceType))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	checkInTime := time.Now()
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
		Code:         b64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s,%s", checkInTime.String(), CUNEXSuccess.RefId)),
	}

	return &responseBody, nil
}

func (s *service) ValidateQRCode(code string) string {
	if code == "" {
		return "Missing URL path parameter 'qrcode'"
	}
	length := 0
	numbers := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for _, r := range code {
		isDigit := false
		for _, num := range numbers {
			if string(r) == num {
				isDigit = true
			}
		}
		if !isDigit {
			return "URL path parameter 'qrcode' contains non-number character(s)"
		}

		length += 1
		if length > 10 {
			break
		}
	}
	if length != 10 {
		return "URL path parameter 'qrcode' must have length of 10"
	}

	return ""
}
