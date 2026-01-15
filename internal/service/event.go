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
	"gorm.io/gorm"
)

type EventService interface {
	PostParticipantService(code string, eventId string, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError)
}

func (s *service) PostParticipantService(code string, eventId string, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError) {
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
		// Get faculty ID from the last 2 digits of student ID
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

	user, getUserErr := s.repo.Event.GetUserForCheckin(ctx, refIdUInt)
	if getUserErr != nil {
		if getUserErr == gorm.ErrRecordNotFound {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Message: "Participant with this ref_id not found",
				Status:  404,
			}
		}

		s.logger.Error().Err(getUserErr).
			Uint64("participant ref_id", refIdUInt).
			Str("function", "EventRepository.GetUserForCheckin")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	event, getEventErr := s.repo.Event.GetEventForCheckin(ctx, eventIdUuid)
	if getEventErr != nil {
		if getEventErr == gorm.ErrRecordNotFound {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Message: "Event with this id not found",
				Status:  404,
			}
		}

		s.logger.Error().Err(getEventErr).
			Str("event_id", eventId).
			Str("function", "EventRepository.GetEventForCheckin")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	status, checkinTime, errCheckStatus := s._CheckCheckinStatus(ctx, eventIdUuid, refIdUInt, string(event.AttendanceType), orgCode, event.EndTime)
	if errCheckStatus != nil {
		return nil, errCheckStatus
	}

	scanRecord := entity.ScanRecordInsert{
		EventID:          eventIdUuid,
		ScannedTimestamp: *checkinTime,
		ParticipantRefID: refIdUInt,
		FirstName:        CUNEXSuccess.FirstName,
		SurName:          CUNEXSuccess.LastName,
		Organization:     CUNEXSuccess.Organization,
		// TODO: get point from front end
		ScannedLocation: entity.Point{X: 0.1, Y: 9.9},
		// TODO: get scanner ID from this user's user_id
		ScannerID: datatypes.NewUUIDv4(),
	}
	rowId, insertErr := s.repo.Event.InsertScanRecord(ctx, &scanRecord)
	if insertErr != nil {
		s.logger.Error().Err(insertErr).
			Uint64("participant ref_id", refIdUInt).
			Str("function", "EventRepository.InsertScanRecord")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	rawCode := fmt.Appendf(nil, "%s.%s", checkinTime.Format("2006-01-02T15:04:05Z"), rowId.String())
	responseBody := dtoRes.GetParticipantRes{
		FirstnameTH:  user.FirstnameTH,
		SurnameTH:    user.SurnameTH,
		TitleTH:      user.TitleTH,
		FirstnameEN:  CUNEXSuccess.FirstName,
		SurnameEN:    CUNEXSuccess.LastName,
		TitleEN:      user.TitleEN,
		RefID:        refIdUInt,
		Organization: CUNEXSuccess.Organization,
		CheckInTime:  *checkinTime,
		Status:       status,
		Code:         b64.StdEncoding.EncodeToString(rawCode),
	}

	return &responseBody, nil
}

func (s *service) _CheckCheckinStatus(ctx context.Context, eventId datatypes.UUID, participantRefId uint64, attendanceType string, orgCode int64, eventEndTime time.Time) (string, *time.Time, *response.APIError) {
	now := time.Now().UTC()

	// Must check if already checked in, regardless of attendance type
	found, err := s.repo.Event.CheckEventParticipation(ctx, eventId, participantRefId)
	if err != nil {
		s.logger.Error().Err(err).
			Uint64("participant ref_id", participantRefId).
			Str("function", "EventRepository.CheckEventParticipation")
		return "", nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}
	if found {
		return string(dtoRes.DUPLICATE), &now, nil
	}

	// If FACULTIES or WHITELIST, must check for access
	if attendanceType == string(entity.FACULTIES) || attendanceType == string(entity.WHITELIST) {
		allow, err := s.repo.Event.CheckEventAccess(ctx, orgCode, participantRefId, attendanceType, eventId)
		if err != nil {
			s.logger.Error().Err(err).
				Uint64("participant ref_id", participantRefId).
				Str("function", "EventRepository.CheckEventAccess")
			return "", nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}

		if !allow {
			return string(dtoRes.FAIL), &now, nil
		}
	}

	// Cannot check in if already past the event's ending time
	if now.After(eventEndTime.UTC()) {
		return string(dtoRes.LATE), &now, nil
	}

	return string(dtoRes.SUCCESS), &now, nil
}
