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
	PostParticipantService(code string, eventId string, userId string, scannedLocX float64, scannedLocY float64, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError)
}

func (s *service) PostParticipantService(code string, eventId string, userId string, scannedLocX float64, scannedLocY float64, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError) {
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

	eventIdErr := uuid.Validate(eventId)
	if eventIdErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Invalid event ID format",
			Status:  400,
		}
	}
	eventIdUuid := datatypes.UUID(datatypes.BinUUIDFromString(eventId))

	userIdErr := uuid.Validate(userId)
	if userIdErr != nil {
		s.logger.Error().Err(userIdErr).
			Str("user_id", userId)
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid user_id from JWT claim",
			Status:  500,
		}
	}
	userIdUuid := datatypes.UUID(datatypes.BinUUIDFromString(userId))

	CUNEXGetQRURL := "https://culab-svc.azurewebsites.net/Service.svc/qrcodeinfo_for_all"
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
			Message: "Failed to perform HTTP request for CU NEX GET qrcode",
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

	switch resp.StatusCode {
	case 200:
		parseErr := json.NewDecoder(resp.Body).Decode(&CUNEXSuccess)
		if parseErr != nil {
			s.logger.Error().Err(parseErr).Str("Error", "Could not parse success response from CU NEX GET qrcode")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Could not parse success response from CU NEX GET qrcode",
				Status:  500,
			}
		}

	case 401:
		// Incorrect ClientId or ClientSecret
		s.logger.Error().Str("Error", "Authorization error from CU NEX GET qrcode")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Authorization error from CU NEX GET qrcode",
			Status:  500,
		}

	case 403:
		// Expired or invalid QR
		return nil, &response.APIError{
			Code:    "GONE",
			Message: "Invalid or expired qrcode",
			Status:  410,
		}

	case 500:
		s.logger.Error().Str("Error", "Internal server error from CU NEX GET qrcode")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal error from CU NEX GET qrcode",
			Status:  500,
		}

	default:
		s.logger.Error().Str("Error", fmt.Sprintf("Response with unexpected status code from CU NEX GET qrcode: %d", resp.StatusCode))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Unknown response from CU NEX GET qrcode",
			Status:  500,
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

	tempCode, err := strconv.ParseUint(CUNEXSuccess.FacultyCode, 10, 8)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid facultyCode returned from CU NEX GET qrcode",
			Status:  500,
		}
	}
	orgCode := uint8(tempCode)

	user, getUserErr := s.repo.Event.GetUserForCheckin(ctx, refIdUInt)
	if getUserErr != nil {
		if getUserErr != gorm.ErrRecordNotFound {
			s.logger.Error().Err(getUserErr).
				Uint64("participant ref_id", refIdUInt).
				Str("function", "EventRepository.GetUserForCheckin")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}

		// Possible for no record; user can be people with no account in our system
		// TODO: API doesn't provide title, so where to get title if this participant isn't in our DB?
		user = &entity.CheckinUserQuery{
			TitleTH: "",
			TitleEN: "",
		}
	}

	event, getEventErr := s.repo.Event.GetEventForCheckin(ctx, eventIdUuid, userIdUuid)
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

	if !event.AllowAllToScan && !event.ThisUserCanScan {
		return nil, &response.APIError{
			Code:    response.ErrForbidden,
			Message: "This user doesn't have permission to be a scanner for this event",
			Status:  403,
		}
	}

	status, checkinTime, errCheckStatus := s._CheckCheckinStatus(ctx, eventIdUuid, refIdUInt, string(event.AttendenceType), orgCode, event.EndTime)
	if errCheckStatus != nil {
		return nil, errCheckStatus
	}

	var (
		orgTH string
		orgEN string
	)
	switch CUNEXSuccess.UserType {
	case entity.STUDENTS:
		orgTH = CUNEXSuccess.FacultyNameTH
		orgEN = CUNEXSuccess.FacultyNameEN

	case entity.STAFFS:
		orgTH = CUNEXSuccess.DepartmentNameTH
		orgEN = CUNEXSuccess.DepartmentNameEN

	default:
		s.logger.Error().Str("Error", fmt.Sprintf("Invalid userType returned from CU NEX GET qrcode: %s", CUNEXSuccess.UserType))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid userType returned from CU NEX GET qrcode",
			Status:  500,
		}
	}

	scanRecord := entity.EventParticipants{
		EventID:          eventIdUuid,
		ScannedTimestamp: *checkinTime,
		ParticipantRefID: refIdUInt,
		FirstName:        CUNEXSuccess.FirstNameEN,
		SurName:          CUNEXSuccess.LastNameEN,
		Organization:     orgEN,
		ScannedLocation:  entity.Point{X: scannedLocX, Y: scannedLocY},
		ScannerID:        &userIdUuid,
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
		FirstnameTH:     nil,
		SurnameTH:       nil,
		TitleTH:         nil,
		FirstnameEN:     nil,
		SurnameEN:       nil,
		TitleEN:         nil,
		RefID:           nil,
		OrganizationTH:  nil,
		OrganizationEN:  nil,
		CheckInTime:     *checkinTime,
		Status:          status,
		Code:            b64.StdEncoding.EncodeToString(rawCode),
		ProfileImageUrl: nil,
	}

	for _, field := range event.RevealedFields {
		switch field {
		case entity.NAME:
			responseBody.FirstnameTH = &CUNEXSuccess.FirstNameTH
			responseBody.FirstnameEN = &CUNEXSuccess.FirstNameEN
			responseBody.TitleTH = &user.TitleTH
			responseBody.SurnameTH = &CUNEXSuccess.LastNameTH
			responseBody.SurnameEN = &CUNEXSuccess.LastNameEN
			responseBody.TitleEN = &user.TitleEN

		case entity.ORGANIZATION:
			responseBody.OrganizationTH = &orgTH
			responseBody.OrganizationEN = &orgEN

		case entity.PHOTO:
			responseBody.ProfileImageUrl = &CUNEXSuccess.ProfileImageUrl

		case entity.REFID:
			responseBody.RefID = &refIdUInt
		}
	}

	return &responseBody, nil
}

func (s *service) _CheckCheckinStatus(ctx context.Context, eventId datatypes.UUID, participantRefId uint64, attendanceType string, orgCode uint8, eventEndTime time.Time) (string, *time.Time, *response.APIError) {
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
