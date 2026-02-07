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
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type EventService interface {
	DeleteById(eventIDStr string, userIDStr string, ctx context.Context) *response.APIError
	DuplicateById(EventID string, userIDStr string, ctx context.Context) (*dtoRes.DuplicateEventRes, *response.APIError)
	Comment(checkInReq dtoReq.CommentReq, ctx context.Context) *response.APIError
	PostParticipantService(code string, eventId string, userId string, scannedLocX float64, scannedLocY float64, ctx context.Context) (*dtoRes.GetParticipantRes, *response.APIError)

	GetOneEventService(eventIdStr string, userIdStr string, ctx context.Context) (res *dtoRes.GetOneEventRes, err *response.APIError)
	GetEventsService(userIDStr string, queryParams map[string]string, ctx context.Context) (*[]dtoRes.GetEventsRes, *response.Pagination, *response.APIError)
}

func (s *service) Comment(commentReq dtoReq.CommentReq, ctx context.Context) *response.APIError {

	decoded, err := base64.StdEncoding.DecodeString(commentReq.EncodedOneTimeCode)
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

	if err := s.repo.Event.Comment(
		checkInRowId,
		timeStamp,
		commentReq.Comment,
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

func (s *service) DeleteById(eventIDStr string, userIDStr string, ctx context.Context) *response.APIError {
	eventID, parseErr := uuid.Parse(eventIDStr)
	if parseErr != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "invalid event_id format",
			Status:  400,
		}
	}

	if _, err := uuid.Parse(userIDStr); err != nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "invalid user_id format",
			Status:  400,
		}
	}

	if eventID == uuid.Nil {
		return &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "nil event_id not allowed",
			Status:  400,
		}
	}

	err := s.repo.Event.DeleteById(eventID, userIDStr, ctx)

	if err != nil {
		logger := s.logger.Error().
			Err(err).
			Str("event_id", eventIDStr).
			Str("user_id", userIDStr).
			Str("action", "delete_event")

		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Msg("event not found")
			return &response.APIError{
				Code:    response.ErrNotFound,
				Message: "event not found",
				Status:  404,
			}
		}

		if errors.Is(err, entity.ErrInsufficientPermissions) {
			logger.Msg("user unauthorized")
			return &response.APIError{
				Code:    response.ErrForbidden,
				Message: "insufficient permissions",
				Status:  403,
			}
		}

		if errors.Is(err, entity.ErrNilUUID) {
			logger.Msg("attempt deleting nil uuid")
			return &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "nil uuid not allowed",
				Status:  400,
			}
		}

		logger.Msg("service failed to delete event")
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	return nil
}

func (s *service) DuplicateById(eventIDStr string, userIDStr string, ctx context.Context) (*dtoRes.DuplicateEventRes, *response.APIError) {
	eventID, parseErr := uuid.Parse(eventIDStr)
	if parseErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "invalid event_id format",
			Status:  400,
		}
	}

	if _, err := uuid.Parse(userIDStr); err != nil {
		return nil, &response.APIError{
			Code:    response.ErrUnauthorized,
			Message: "invalid user_id format",
			Status:  401,
		}
	}

	isMember, authErr := s.repo.Event.IsUserEventAdmin(eventID, userIDStr, ctx)
	if authErr != nil {
		s.logger.Error().
			Err(authErr).
			Str("event_id", eventIDStr).
			Str("user_id", userIDStr).
			Str("action", "duplicate_event_auth").
			Msg("failed to check event permissions")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	if !isMember {
		return nil, &response.APIError{
			Code:    response.ErrForbidden,
			Message: "insufficient permissions",
			Status:  403,
		}
	}

	originalEvent, findErr := s.repo.Event.FindById(eventID, ctx)
	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Message: "event not found",
				Status:  404,
			}
		}
		s.logger.Error().
			Err(findErr).
			Str("event_id", eventIDStr).
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
			Str("event_id", eventIDStr).
			Str("action", "duplicate_event").
			Msg("failed to duplicate event")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to duplicate event",
			Status:  500,
		}
	}

	return &dtoRes.DuplicateEventRes{
		DuplicatedEventId: createdEvent.ID.String(),
	}, nil
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

func (s *service) GetOneEventService(eventIdStr string, userIdStr string, ctx context.Context) (*dtoRes.GetOneEventRes, *response.APIError) {
	eventIdErr := uuid.Validate(eventIdStr)
	if eventIdErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Invalid URL path parameter 'id'",
			Status:  400,
		}
	}
	eventId := datatypes.UUID(datatypes.BinUUIDFromString(eventIdStr))

	userIdErr := uuid.Validate(userIdStr)
	if userIdErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Invalid user_id from JWT claim",
			Status:  500,
		}
	}
	userId := datatypes.UUID(datatypes.BinUUIDFromString(userIdStr))

	eventWithCount, agenda, err := s.repo.Event.GetOneEvent(eventId, userId, ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Message: "Event with this id not found",
				Status:  404,
			}
		} else {
			s.logger.Error().Err(err).
				Str("function", "EventRepository.GetOneEvent")
			return nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error",
				Status:  500,
			}
		}
	}

	agendaDTO := []dtoRes.GetOneEventAgenda{}
	if len(*agenda) > 0 {
		for _, slot := range *agenda {
			agendaDTO = append(agendaDTO, dtoRes.GetOneEventAgenda{
				ActivityName: slot.ActivityName,
				StartTime:    slot.StartTime.UTC(),
				EndTime:      slot.EndTime.UTC(),
			})
		}
	}

	finalRes := dtoRes.GetOneEventRes{
		Name:            eventWithCount.Name,
		Organizer:       eventWithCount.Organizer,
		Description:     eventWithCount.Description,
		StartTime:       eventWithCount.StartTime.UTC(),
		EndTime:         eventWithCount.EndTime.UTC(),
		Location:        eventWithCount.Location,
		TotalRegistered: eventWithCount.TotalRegistered,
		EvaluationForm:  eventWithCount.EvaluationForm,
		Agenda:          agendaDTO,
		Role:            eventWithCount.Role,
	}

	return &finalRes, nil
}

func (s *service) GetEventsService(userIDStr string, queryParams map[string]string, ctx context.Context) (*[]dtoRes.GetEventsRes, *response.Pagination, *response.APIError) {
	uuidValidationErr := uuid.Validate(userIDStr)
	if uuidValidationErr != nil {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Invalid UUID format for user_id from middleware",
			Status:  500,
		}
	}
	userID := datatypes.UUID(datatypes.BinUUIDFromString(userIDStr))

	pageQuery, pageOk := queryParams["page"]
	var page int
	if pageOk {
		pageInt, err := strconv.Atoi(pageQuery)
		if err != nil {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'page' must be int",
				Status:  400,
			}
		}
		if pageInt < 0 {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'page' must be greater than 0",
				Status:  400,
			}
		}
		page = pageInt
	}

	size := 8
	sizeQuery, sizeOk := queryParams["pageSize"]
	if sizeOk {
		pageSizeInt, err := strconv.Atoi(sizeQuery)
		if err != nil {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'pageSize' must be int",
				Status:  400,
			}
		}
		if pageSizeInt < 1 || pageSizeInt > 10 {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'pageSize' must be within range [1, 10]",
				Status:  400,
			}
		}
		size = pageSizeInt
	}

	search := ""
	searchQuery, searchOk := queryParams["search"]
	if searchOk {
		search = strings.TrimSpace(searchQuery)
		if utf8.RuneCountInString(search) > 256 {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'search' longer than 256 characters",
				Status:  400,
			}
		}
	}

	formattedRes := []dtoRes.GetEventsRes{}

	managedQuery, managedOk := queryParams["managed"]
	// 'managed' not present -> get discovery events
	if !managedOk {
		if !pageOk {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "Missing required URL query parameter: page",
				Status:  400,
			}
		}
		res, total, hasNext, err := s.repo.Event.GetDiscoveryEvents(userID, page, size, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userIDStr).
				Str("function", "EventRepository.GetDiscoveryEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting discovery events",
				Status:  500,
			}
		}
		s._GetEventsDTOFormat(res, &formattedRes)
		return &formattedRes, &response.Pagination{
			Page:     page,
			PageSize: size,
			Total:    total,
			HasNext:  hasNext,
		}, nil
	}

	// 'managed' present -> parse and get managed or participated events
	managed, err := strconv.ParseBool(managedQuery)
	if err != nil {
		return nil, nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "URL query parameter 'managed' must be boolean",
			Status:  400,
		}
	}
	switch managed {
	case true:
		res, err := s.repo.Event.GetManagedEvents(userID, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userIDStr).
				Str("function", "EventRepository.GetManagedEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting managed events",
				Status:  500,
			}
		}
		s._GetEventsDTOFormat(res, &formattedRes)
		return &formattedRes, nil, nil

	default:
		if !pageOk {
			return nil, nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "Missing required URL query parameter: page",
				Status:  400,
			}
		}
		res, total, hasNext, err := s.repo.Event.GetAttendedEvents(userID, page, size, search, ctx)
		if err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userIDStr).
				Str("function", "EventRepository.GetAttendedEvents").
				Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
			return nil, nil, &response.APIError{
				Code:    response.ErrInternalError,
				Message: "Internal DB error on getting attended events",
				Status:  500,
			}
		}
		s._GetEventsDTOFormat(res, &formattedRes)
		return &formattedRes, &response.Pagination{
			Page:     page,
			PageSize: size,
			Total:    total,
			HasNext:  hasNext,
		}, nil
	}
}

func (s *service) _GetEventsDTOFormat(rawResult *[]entity.GetEventsQueryResult, result *[]dtoRes.GetEventsRes) {
	length := len(*rawResult)
	if length > 0 {
		for i := 0; i < length; i++ {
			*result = append(*result, dtoRes.GetEventsRes{
				ID:             (*rawResult)[i].ID.String(),
				Name:           (*rawResult)[i].Name,
				Organizer:      (*rawResult)[i].Organizer,
				Description:    (*rawResult)[i].Description,
				StartTime:      (*rawResult)[i].StartTime.UTC(),
				EndTime:        (*rawResult)[i].EndTime.UTC(),
				Location:       (*rawResult)[i].Location,
				Role:           (*rawResult)[i].Role,
				EvaluationForm: (*rawResult)[i].EvaluationForm,
			})
		}
	}
}
