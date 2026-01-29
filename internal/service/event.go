package service

import (
	"context"
	"encoding/base64"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type EventService interface {
	DeleteById(EventID string, ctx context.Context) *response.APIError
	DuplicateById(EventID string, ctx context.Context) (*entity.Event, *response.APIError)
	CheckIn(checkInReq dtoReq.CheckInReq, ctx context.Context) *response.APIError
	PostParticipantService(code string, eventId string, userId string, scannedLocX float64, scannedLocY float64, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError)
}

func (s *service) CheckIn(checkInReq dtoReq.CheckInReq, ctx context.Context) *response.APIError {

	decoded, err := base64.StdEncoding.DecodeString(checkInReq.EncodedOneTimeCode)
	if err != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "failed to interpret one_time_code as base64 encoded",
			Status:  400,
		}
	}

	raw := string(decoded)
	idx := strings.LastIndex(raw, ".")
	if idx == -1 {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "invalid one_time_code format",
			Status:  400,
		}
	}

	strTimeStamp := raw[:idx]
	strCheckInRowId := raw[idx+1:]

	checkInRowId, err := uuid.Parse(strCheckInRowId)
	if err != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "failed to parse check-in target id",
			Status:  400,
		}
	}

	timeStamp, err := time.Parse(time.RFC3339, strTimeStamp)
	if err != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "failed to parse timeStamp to go time",
			Status:  400,
		}
	}

	s.logger.Info().
		Str("timeStamp", timeStamp.String()).
		Str("checkInRowId", checkInRowId.String()).
		Msg("Received timeStamp and target row-id to check-in Event-Participant")

	if err := s.repo.Event.CheckIn(
		checkInRowId,
		timeStamp,
		checkInReq.Comment,
		ctx,
	); err != nil {

		if errors.Is(err, entity.ErrAlreadyCheckedIn) {
			return &response.APIError{
				Code:    response.ErrConflict,
				Message: err.Error(),
				Status:  409,
			}
		}

		if errors.Is(err, entity.ErrCheckInTargetNotFound) {
			return &response.APIError{
				Code:    response.ErrBadRequest,
				Message: err.Error(),
				Status:  400,
			}
		}

		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	return nil
}

func (s *service) DeleteById(EventId string, ctx context.Context) *response.APIError {
	event_id, parseErr := uuid.Parse(EventId)
	if parseErr != nil {
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to parse event_id to uuid",
			Status:  500,
		}
	}

	eventDeleteErr := s.repo.Event.DeleteById(event_id, ctx)

	if errors.Is(eventDeleteErr, gorm.ErrRecordNotFound) {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("event not found")
		return &response.APIError{
			Code:    response.ErrNotFound,
			Message: "event not found",
			Status:  404,
		}
	}

	if errors.Is(eventDeleteErr, entity.ErrNilUUID) {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("attempt deleting nil uuid")
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "nil uuid not allowed",
			Status:  400,
		}
	}

	if eventDeleteErr != nil {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("service failed to delete event")
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	return nil
}

func (s *service) DuplicateById(EventId string, ctx context.Context) (*entity.Event, *response.APIError) {
	event_id, parseErr := uuid.Parse(EventId)
	if parseErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to parse event_id to uuid",
			Status:  400,
		}
	}

	originalEvent, findErr := s.repo.Event.FindById(event_id, ctx)
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "specified event not found",
				Status:  400,
			}
		}
		s.logger.Error().
			Err(findErr).
			Str("event_id", EventId).
			Str("action", "duplicate_event_find").
			Msg("failed to find event for duplication")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	newEvent := *originalEvent
	newEvent.ID = datatypes.UUID(uuid.New())

	// breaking the memory link from originalEvent
	// Whitelist
	newEvent.EventWhitelist = make([]entity.EventWhitelist, 0, len(originalEvent.EventWhitelist))
	for _, item := range originalEvent.EventWhitelist {
		newEvent.EventWhitelist = append(newEvent.EventWhitelist, entity.EventWhitelist{
			AttendeeRefID: item.AttendeeRefID,
		})
	}

	// Faculties
	newEvent.EventAllowedFaculties = make([]entity.EventAllowedFaculties, 0, len(originalEvent.EventAllowedFaculties))
	for _, item := range originalEvent.EventAllowedFaculties {
		newEvent.EventAllowedFaculties = append(newEvent.EventAllowedFaculties, entity.EventAllowedFaculties{
			FacultyNO: item.FacultyNO,
		})
	}

	// Agenda
	newEvent.EventAgenda = make([]entity.EventAgenda, 0, len(originalEvent.EventAgenda))
	for _, item := range originalEvent.EventAgenda {
		newEvent.EventAgenda = append(newEvent.EventAgenda, entity.EventAgenda{
			ActivityName: item.ActivityName,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
		})
	}

	createdEvent, createErr := s.repo.Event.Create(&newEvent, ctx)
	if createErr != nil {
		s.logger.Error().
			Err(createErr).
			Str("event_id", EventId).
			Str("action", "duplicate_event").
			Msg("failed to duplicate event")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to duplicate event",
			Status:  500,
		}
	}

	return createdEvent, nil
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

	// Request for participant profile
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
			Code:    response.ErrNotFound,
			Message: "qrcode not found (expired or invalid)",
			Status:  404,
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

	// Format participant info from CU NEX API to fit our uses
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

	// Insert participant now to allow inserting them into EventParticipants later
	userToInsert := entity.User{
		RefID:       refIdUInt,
		FirstnameTH: CUNEXSuccess.FirstNameTH,
		SurnameTH:   CUNEXSuccess.LastNameTH,
		TitleTH:     "",
		FirstnameEN: CUNEXSuccess.FirstNameEN,
		SurnameEN:   CUNEXSuccess.LastNameEN,
		TitleEN:     "",
	}
	user, createUserErr := s.CreateUserIfNotExists(&userToInsert, ctx)
	if createUserErr != nil {
		return nil, createUserErr
	}

	// Get event info for checking scanning/check in permission
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

	status, checkinTime, rowId, errCheckStatus := s.CheckCheckinStatus(ctx, eventIdUuid, user.RefID, user.ID, string(event.AttendenceType), orgCode, event.EndTime)
	if errCheckStatus != nil {
		return nil, errCheckStatus
	}

	// Format org before proceeding with steps according to status
	// to make EventParticipants insertion possible
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

	switch status {
	case string(dtoRes.FAIL):
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Failed to check in to the event (late or no permission).",
			Status:  400,
		}

	case string(dtoRes.SUCCESS):
		scanRecord := entity.EventParticipants{
			EventID:          eventIdUuid,
			ScannedTimestamp: *checkinTime,
			ParticipantID:    user.ID,
			Organization:     orgEN,
			ScannedLocation:  entity.Point{X: scannedLocX, Y: scannedLocY},
			ScannerID:        &userIdUuid,
		}
		rowIdInsert, insertErr := s.repo.Event.InsertScanRecord(ctx, &scanRecord)
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

		rowId = rowIdInsert
	}

	raw := fmt.Appendf(nil, "%s.%s", checkinTime.Format(time.RFC3339Nano), rowId.String())
	checkInCode := b64.StdEncoding.EncodeToString(raw)

	// Finally, format response according to revealed_fields of this event
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
		Code:            checkInCode,
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
			temp := s.FormatRefIdToStr(refIdUInt)
			responseBody.RefID = &temp
		}
	}

	return &responseBody, nil
}

// returns (status, checkInTime, rowId (if duplication found), error)
func (s *service) CheckCheckinStatus(ctx context.Context, eventId datatypes.UUID, participantRefId uint64, participantId datatypes.UUID, attendanceType string, orgCode uint8, eventEndTime time.Time) (string, *time.Time, *datatypes.UUID, *response.APIError) {
	now := time.Now().UTC()

	// Must check if already checked in, regardless of attendance type
	rowId, err := s.repo.Event.CheckEventParticipation(ctx, eventId, participantId)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error().Err(err).
			Str("participant id", participantId.String()).
			Str("function", "EventRepository.CheckEventParticipation")
		return "", nil, nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}
	if rowId != nil {
		return string(dtoRes.DUPLICATE), &now, rowId, nil
	}

	// If FACULTIES or WHITELIST, must check for access
	if attendanceType == string(entity.FACULTIES) || attendanceType == string(entity.WHITELIST) {
		allow, err := s.repo.Event.CheckEventAccess(ctx, orgCode, participantRefId, attendanceType, eventId)
		if err != nil {
			s.logger.Error().Err(err).
				Uint64("participant ref_id", participantRefId).
				Str("function", "EventRepository.CheckEventAccess")
			return "", nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}

		if !allow {
			return string(dtoRes.FAIL), &now, nil, nil
		}
	}

	// Cannot check in if already past the event's ending time
	if now.After(eventEndTime.UTC()) {
		return string(dtoRes.FAIL), &now, nil, nil
	}

	return string(dtoRes.SUCCESS), &now, nil, nil
}
