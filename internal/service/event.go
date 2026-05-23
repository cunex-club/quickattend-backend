package service

import (
	"context"
	"encoding/base64"
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
	"github.com/cunex-club/quickattend-backend/internal/repository"
)

var thaiLoc = time.FixedZone(entity.ThaiTZ, 7*3600)

type EventService interface {
	DeleteById(eventIDStr string, userIDStr string, ctx context.Context) *response.APIError
	DuplicateById(Req dtoReq.DuplicateEventReq, EventID string, userIDStr string, ctx context.Context) (*dtoRes.DuplicateEventRes, *response.APIError)
	Comment(checkInReq dtoReq.CommentReq, ctx context.Context) *response.APIError
	PostParticipantService(code string, eventId string, userId string, scannedLocX float64, scannedLocY float64, ctx context.Context) (*dtoRes.PostParticipantRes, *response.APIError)

	GetOneEventService(eventIdStr string, userIdStr string, ctx context.Context) (res *dtoRes.GetOneEventRes, err *response.APIError)

	CreateEvent(ctx context.Context, req dtoReq.CreateEventReq, userId string) (*dtoRes.CreateEventRes, error)
	UpdateEvent(ctx context.Context, id string, userId string, updates dtoReq.UpdateEventReq) (*dtoRes.UpdateEventRes, error)

	GetEventsValidateArgs(userIDStr string, queryParams map[string]string, ctx context.Context) (validated *GetEventsValidateArgsReturn, err *response.APIError)
	GetMyEventsService(userID datatypes.UUID, search string, ctx context.Context) (res *[]dtoRes.GetEventsRes, err *response.APIError)
	GetDiscoveryEventsService(args *GetEventsWithPaginationArgs) (res *[]dtoRes.GetDiscoveryEventsRes, pagination *response.Pagination, err *response.APIError)
	GetPastEventsService(args *GetEventsWithPaginationArgs) (res *[]dtoRes.GetEventsRes, pagination *response.Pagination, err *response.APIError)
}

type GetEventsValidateArgsReturn struct {
	UserID   datatypes.UUID
	MyEvents GetEventsMode
	Page     int
	PageSize int
	Search   string
}

type GetEventsWithPaginationArgs struct {
	UserID   datatypes.UUID
	Page     int
	PageSize int
	Search   string
	Ctx      context.Context
}

type GetEventsMode int

const (
	MyEvents GetEventsMode = iota
	PastEvents
	Discovery
)

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

		if errors.Is(err, entity.ErrAlreadyCommented) {
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

func (s *service) DuplicateById(Req dtoReq.DuplicateEventReq, eventIDStr string, userIDStr string, ctx context.Context) (*dtoRes.DuplicateEventRes, *response.APIError) {
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

	isOwner, authErr := s.repo.Event.IsUserEventOwner(eventID, userIDStr, ctx)
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

	if !isOwner {
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

	// validate payload
	// timezone
	if err := validateThaiTimezone(Req.Timezone); err != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Status:  400,
			Message: err.Error(),
		}
	}

	// start_time, end_time
	startTime, err := parseTime(Req.StartTime)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Status:  400,
			Message: err.Error(),
		}
	}
	endTime, err := parseTime(Req.EndTime)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Status:  400,
			Message: err.Error(),
		}
	}
	if !endTime.After(startTime) {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Status:  400,
			Message: "end_time must be after start_time",
		}
	}
	if !isSameDay(startTime, endTime) {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Status:  400,
			Message: fmt.Sprintf("start_time and end_time must be on the same day in timezone %s", entity.ThaiTZ),
		}
	}

	// agenda
	agendas, err := buildAgendas(Req.Agenda, startTime, endTime)
	if err != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Status:  400,
			Message: err.Error(),
		}
	}

	// build new event
	newEvent := *originalEvent
	newEvent.ID = datatypes.UUID(uuid.New())

	// break the memory link from originalEvent, insert new information
	// location
	newEvent.Location = Req.Location

	// start_time
	newEvent.StartTime = startTime

	// end_time
	newEvent.EndTime = endTime

	// evaluation_form
	newEvent.EvaluationForm = Req.EvaluationForm

	// location_point
	newEvent.LocationPoint = entity.Point{
		X: Req.LocationLong,
		Y: Req.LocationLat,
	}

	// agenda
	newEvent.EventAgenda = make([]entity.EventAgenda, 0, len(agendas))
	for _, item := range agendas {
		newEvent.EventAgenda = append(newEvent.EventAgenda, entity.EventAgenda{
			ActivityName: item.ActivityName,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
		})
	}

	// copy over other existing info
	// Whitelist
	newEvent.EventWhitelist = make([]entity.EventWhitelist, 0, len(originalEvent.EventWhitelist))
	for _, item := range originalEvent.EventWhitelist {
		newEvent.EventWhitelist = append(newEvent.EventWhitelist, entity.EventWhitelist{
			AttendeeRefID: item.AttendeeRefID,
		})
	}
	newEvent.EventWhitelistPending = make([]entity.EventWhitelistPending, 0, len(originalEvent.EventWhitelistPending))
	for _, item := range originalEvent.EventWhitelistPending {
		newEvent.EventWhitelistPending = append(newEvent.EventWhitelistPending, entity.EventWhitelistPending{
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

	// Users
	newEvent.EventUser = make([]entity.EventUser, 0, len(originalEvent.EventUser))
	for _, item := range originalEvent.EventUser {
		newEvent.EventUser = append(newEvent.EventUser, entity.EventUser{
			UserID: item.UserID,
			Role:   item.Role,
		})
	}
	newEvent.EventUserPending = make([]entity.EventUserPending, 0, len(originalEvent.EventUserPending))
	for _, item := range originalEvent.EventUserPending {
		newEvent.EventUserPending = append(newEvent.EventUserPending, entity.EventUserPending{
			UserRefID: item.UserRefID,
			Role:      item.Role,
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

func (s *service) PostParticipantService(code string, eventId string, userId string, scannedLocX float64, scannedLocY float64, ctx context.Context) (*dtoRes.PostParticipantRes, *response.APIError) {
	if code == "" {
		return nil, &response.APIError{
			Code:    "INVALID_QR",
			Message: "Missing URL path parameter 'qrcode'",
			Status:  400,
		}
	}
	if len(code) != 10 {
		return nil, &response.APIError{
			Code:    "INVALID_QR",
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
				Code:    "INVALID_QR",
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
			Code:    response.ErrBadRequest,
			Message: "Invalid format of user_id from JWT claim",
			Status:  400,
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

	req, formNewReqErr := http.NewRequest(http.MethodGet, CUNEXGetQRURL, nil)
	if formNewReqErr != nil {
		s.logger.Error().Err(formNewReqErr).Str("Error", "Failed to form new HTTP request for CU NEX GET qrcode")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Failed to form HTTP request for CU NEX GET qrcode",
			Status:  500,
		}
	}

	query := req.URL.Query()
	query.Add("qrcode", code)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ClientId", clientId)
	req.Header.Set("ClientSecret", clientSecret)

	resp, doErr := s.httpClient.Do(req)
	if doErr != nil {
		s.logger.Error().Err(doErr).Msg("Failed to perform request for CU NEX GET qrcode")
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Failed to perform request for CU NEX GET qrcode",
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
			Code:    "INVALID_QR",
			Message: "qrcode expired or invalid",
			Status:  400,
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
			Message: "Response with unknown status code from CU NEX GET qrcode",
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

	userToUpsert := entity.User{
		RefID:           refIdUInt,
		FirstnameTH:     CUNEXSuccess.FirstNameTH,
		SurnameTH:       CUNEXSuccess.LastNameTH,
		FirstnameEN:     CUNEXSuccess.FirstNameEN,
		SurnameEN:       CUNEXSuccess.LastNameEN,
		FacultyNameTH:   orgTH,
		FacultyNameEN:   orgEN,
		ProfileImageURL: CUNEXSuccess.ProfileImageUrl,
	}
	notToUpdate := []string{"title_th", "title_en"}
	user, upsertErr := s.repo.Auth.UpsertUserByRefId(&userToUpsert, &notToUpdate, ctx)
	if upsertErr != nil {
		s.logger.Error().Err(upsertErr).
			Uint64("participant_ref_id", refIdUInt).
			Str("action", "upsert_user_by_ref_id").
			Msg("failed to upsert user by ref id")

		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error",
			Status:  500,
		}
	}

	// Get event info for checking scanning/check in permission
	event, getEventErr := s.repo.Event.GetEventForCheckin(ctx, eventIdUuid, userIdUuid)
	if getEventErr != nil {
		if getEventErr == gorm.ErrRecordNotFound {
			return nil, &response.APIError{
				Code:    "EVENT_NOT_FOUND",
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
			Code:    "SCANNER_NO_PERMISSION",
			Message: "This user doesn't have permission to be a scanner for this event",
			Status:  403,
		}
	}

	status, checkinTime, rowId, errCheckStatus := s.CheckCheckinStatus(ctx, eventIdUuid, user.RefID, user.ID, string(event.AttendenceType), orgCode, event.EndTime)
	if errCheckStatus != nil {
		return nil, errCheckStatus
	}

	switch status {
	case string(dtoRes.FAIL):
		return nil, &response.APIError{
			Code:    "PARTICIPANT_NO_PERMISSION",
			Message: "Failed to check in to the event (late or no permission).",
			Status:  403,
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
	checkInCode := base64.StdEncoding.EncodeToString(raw)

	// Finally, format response according to revealed_fields of this event
	responseBody := dtoRes.PostParticipantRes{
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
		case entity.ParticipantName:
			responseBody.FirstnameTH = &CUNEXSuccess.FirstNameTH
			responseBody.FirstnameEN = &CUNEXSuccess.FirstNameEN
			responseBody.TitleTH = &user.TitleTH
			responseBody.SurnameTH = &CUNEXSuccess.LastNameTH
			responseBody.SurnameEN = &CUNEXSuccess.LastNameEN
			responseBody.TitleEN = &user.TitleEN

		case entity.ParticipantOrganization:
			responseBody.OrganizationTH = &orgTH
			responseBody.OrganizationEN = &orgEN

		case entity.ParticipantPhoto:
			responseBody.ProfileImageUrl = &CUNEXSuccess.ProfileImageUrl

		case entity.ParticipantRefID:
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
	if attendanceType == string(entity.AttendanceFaculties) || attendanceType == string(entity.AttendanceWhitelist) {
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

	result, err := s.repo.Event.GetOneEvent(eventId, userId, ctx)
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

	// Format each attribute

	usersDTO := make([]dtoRes.GetOneEventUser, 0, len(result.EventUser))
	if len(result.EventUser) > 0 {
		for _, user := range result.EventUser {
			u := user.User
			usersDTO = append(usersDTO, dtoRes.GetOneEventUser{
				RefID:           s.FormatRefIdToStr(u.RefID),
				FirstnameTH:     u.FirstnameTH,
				SurnameTH:       u.SurnameTH,
				TitleTH:         u.TitleTH,
				FacultyNameTH:   u.FacultyNameTH,
				FirstnameEN:     u.FirstnameEN,
				SurnameEN:       u.SurnameEN,
				TitleEN:         u.TitleEN,
				FacultyNameEN:   u.FacultyNameEN,
				ProfileImageURL: u.ProfileImageURL,
				Role:            string(user.Role),
			})
		}
	}

	usersPendingDTO := make([]dtoRes.GetOneEventUserPending, 0, len(result.EventUserPending))
	if len(result.EventUserPending) > 0 {
		for _, user := range result.EventUserPending {
			usersPendingDTO = append(usersPendingDTO, dtoRes.GetOneEventUserPending{
				RefID: s.FormatRefIdToStr(user.UserRefID),
				Role:  string(user.Role),
			})
		}
	}

	agendaDTO := make([]dtoRes.GetOneEventAgenda, 0, len(result.EventAgenda))
	if len(result.EventAgenda) > 0 {
		for _, slot := range result.EventAgenda {
			agendaDTO = append(agendaDTO, dtoRes.GetOneEventAgenda{
				ActivityName: slot.ActivityName,
				StartTime:    slot.StartTime.UTC(),
				EndTime:      slot.EndTime.UTC(),
			})
		}
	}

	allowedFacDTO := make([]dtoRes.GetOneEventAllowedFaculties, 0, len(result.EventAllowedFaculties))
	if len(result.EventAllowedFaculties) > 0 {
		for _, faculty := range result.EventAllowedFaculties {
			allowedFacDTO = append(allowedFacDTO, dtoRes.GetOneEventAllowedFaculties{
				FacultyNO: faculty.FacultyNO,
			})
		}
	}

	whitelistDTO := make([]dtoRes.GetOneEventWhitelist, 0, len(result.EventWhitelist))
	if len(result.EventWhitelist) > 0 {
		for _, wl := range result.EventWhitelist {
			wlUser := wl.User
			whitelistDTO = append(whitelistDTO, dtoRes.GetOneEventWhitelist{
				RefID:           s.FormatRefIdToStr(wlUser.RefID),
				FirstnameTH:     wlUser.FirstnameTH,
				SurnameTH:       wlUser.SurnameTH,
				TitleTH:         wlUser.TitleTH,
				FacultyNameTH:   wlUser.FacultyNameTH,
				FirstnameEN:     wlUser.FirstnameEN,
				SurnameEN:       wlUser.SurnameEN,
				TitleEN:         wlUser.TitleEN,
				FacultyNameEN:   wlUser.FacultyNameEN,
				ProfileImageURL: wlUser.ProfileImageURL,
			})
		}
	}

	whitelistPendingDTO := make([]dtoRes.GetOneEventWhitelistPending, 0, len(result.EventWhitelistPending))
	if len(result.EventWhitelistPending) > 0 {
		for _, wl := range result.EventWhitelistPending {
			whitelistPendingDTO = append(whitelistPendingDTO, dtoRes.GetOneEventWhitelistPending{
				RefID: s.FormatRefIdToStr(wl.AttendeeRefID),
			})
		}
	}

	revealedFields := make([]string, 0, len(result.RevealedFields))
	for _, field := range result.RevealedFields {
		if field != "" {
			revealedFields = append(revealedFields, string(field))
		}
	}
	finalRes := dtoRes.GetOneEventRes{
		Name:             result.Name,
		Organizer:        result.Organizer,
		Description:      result.Description,
		StartTime:        result.StartTime.UTC(),
		EndTime:          result.EndTime.UTC(),
		Location:         result.Location,
		LocationLat:      result.LocationPoint.Y,
		LocationLong:     result.LocationPoint.X,
		TotalRegistered:  result.TotalRegistered,
		EvaluationForm:   result.EvaluationForm,
		AllowAllToScan:   result.AllowAllToScan,
		RevealedFields:   revealedFields,
		AttendanceType:   result.AttendenceType,
		Role:             result.Role,
		Agenda:           agendaDTO,
		User:             usersDTO,
		UserPending:      usersPendingDTO,
		AllowedFaculties: allowedFacDTO,
		WhiteList:        whitelistDTO,
		WhiteListPending: whitelistPendingDTO,
	}

	return &finalRes, nil
}

func (s *service) GetEventsValidateArgs(userIDStr string, queryParams map[string]string, ctx context.Context) (validated *GetEventsValidateArgsReturn, err *response.APIError) {
	uuidValidationErr := uuid.Validate(userIDStr)
	if uuidValidationErr != nil {
		return nil, &response.APIError{
			Code:    response.ErrBadRequest,
			Message: "Invalid UUID format for user_id from middleware",
			Status:  500,
		}
	}
	userID := datatypes.UUID(datatypes.BinUUIDFromString(userIDStr))

	// // User must exist
	_, userErr := s.repo.Auth.GetUserById(userID, ctx)
	if userErr != nil {
		if userErr == gorm.ErrRecordNotFound {
			return nil, &response.APIError{
				Code:    response.ErrNotFound,
				Message: "User not found",
				Status:  404,
			}
		}

		s.logger.Error().Err(userErr).
			Str("user_id", userIDStr).
			Str("function", "AuthRepository.GetUserById").
			Msg(fmt.Sprintf("Internal DB error: %s", userErr.Error()))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error on getting user",
			Status:  500,
		}
	}

	pageQuery, pageOk := queryParams["page"]
	var page int
	if pageOk {
		pageInt, err := strconv.Atoi(pageQuery)
		if err != nil {
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'page' must be int",
				Status:  400,
			}
		}
		if pageInt < 0 {
			return nil, &response.APIError{
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
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'pageSize' must be int",
				Status:  400,
			}
		}
		if pageSizeInt < 1 || pageSizeInt > 10 {
			return nil, &response.APIError{
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
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'search' longer than 256 characters",
				Status:  400,
			}
		}
	}

	var mode GetEventsMode
	myeventsQuery, myeventsOk := queryParams["myevents"]
	if !myeventsOk {
		// 'myevents' not present -> get discovery events, which requires 'page'
		if !pageOk {
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "Missing URL query parameter 'page'; required when 'myevents' is false or not given",
				Status:  400,
			}
		}
		mode = Discovery

	} else {
		myevents, myeventsErr := strconv.ParseBool(myeventsQuery)
		if myeventsErr != nil {
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "URL query parameter 'myevents' must be boolean",
				Status:  400,
			}
		}

		if !myevents && !pageOk {
			// 'myevents' false -> get Past Events, which requires 'page'
			return nil, &response.APIError{
				Code:    response.ErrBadRequest,
				Message: "Missing URL query parameter 'page'; required when 'myevents' is false or not given",
				Status:  400,
			}
		}

		if myevents {
			mode = MyEvents
		} else {
			mode = PastEvents
		}
	}

	return &GetEventsValidateArgsReturn{
		UserID:   userID,
		MyEvents: mode,
		Page:     page,
		PageSize: size,
		Search:   search,
	}, nil
}

func (s *service) GetMyEventsService(userID datatypes.UUID, search string, ctx context.Context) (*[]dtoRes.GetEventsRes, *response.APIError) {
	args := repository.GetEventsArguments{
		UserID: userID,
		Search: search,
		Ctx:    ctx,
	}

	res, err := s.repo.Event.GetMyEvents(&args)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", userID.String()).
			Str("function", "EventRepository.GetMyEvents").
			Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
		return nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error on getting My Events",
			Status:  500,
		}
	}

	final := make([]dtoRes.GetEventsRes, 0, len(*res))
	s.getEventsDTOFormat(res, &final)
	return &final, nil
}

func (s *service) GetPastEventsService(args *GetEventsWithPaginationArgs) (*[]dtoRes.GetEventsRes, *response.Pagination, *response.APIError) {
	repoArgs := repository.GetEventsArguments{
		UserID:   args.UserID,
		Page:     args.Page,
		PageSize: args.PageSize,
		Search:   args.Search,
		Ctx:      args.Ctx,
	}

	res, total, hasNext, err := s.repo.Event.GetPastEvents(&repoArgs)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", args.UserID.String()).
			Str("function", "EventRepository.GetPastEvents").
			Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
		return nil, nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error on getting Past Events",
			Status:  500,
		}
	}

	final := make([]dtoRes.GetEventsRes, 0, len(*res))
	s.getEventsDTOFormat(res, &final)
	return &final, &response.Pagination{
		Page:     args.Page,
		PageSize: args.PageSize,
		Total:    total,
		HasNext:  hasNext,
	}, nil
}

func (s *service) GetDiscoveryEventsService(args *GetEventsWithPaginationArgs) (*[]dtoRes.GetDiscoveryEventsRes, *response.Pagination, *response.APIError) {
	repoArgs := repository.GetEventsArguments{
		UserID:   args.UserID,
		Page:     args.Page,
		PageSize: args.PageSize,
		Search:   args.Search,
		Ctx:      args.Ctx,
	}

	res, total, hasNext, err := s.repo.Event.GetDiscoveryEvents(&repoArgs)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", args.UserID.String()).
			Str("function", "EventRepository.GetDiscoveryEvents").
			Msg(fmt.Sprintf("Internal DB error: %s", err.Error()))
		return nil, nil, &response.APIError{
			Code:    response.ErrInternalError,
			Message: "Internal DB error on getting discovery events",
			Status:  500,
		}
	}

	final := make([]dtoRes.GetDiscoveryEventsRes, 0, len(*res))
	s.getDiscoveryEventsDTOFormat(res, &final)
	return &final, &response.Pagination{
		Page:     args.Page,
		PageSize: args.PageSize,
		Total:    total,
		HasNext:  hasNext,
	}, nil
}

func (s *service) getEventsDTOFormat(rawResult *[]entity.GetEventsQueryResult, result *[]dtoRes.GetEventsRes) {
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

func (s *service) getDiscoveryEventsDTOFormat(rawResult *[]entity.GetDiscoveryEvents, result *[]dtoRes.GetDiscoveryEventsRes) {
	deref := *rawResult
	if len(deref) > 0 {
		for i := 0; i < len(deref); i++ {
			*result = append(*result, dtoRes.GetDiscoveryEventsRes{
				ID:             deref[i].ID.String(),
				Name:           deref[i].Name,
				Organizer:      deref[i].Organizer,
				Description:    deref[i].Description,
				StartTime:      deref[i].StartTime.UTC(),
				EndTime:        deref[i].EndTime.UTC(),
				Location:       deref[i].Location,
				EvaluationForm: deref[i].EvaluationForm,
				LocationLat:    deref[i].LocationPoint.Y,
				LocationLong:   deref[i].LocationPoint.X,
			})
		}
	}
}

func (s *service) CreateEvent(ctx context.Context, req dtoReq.CreateEventReq, userId string) (*dtoRes.CreateEventRes, error) {
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, errors.New("Invalid user id")
	}

	// Validate that the requester is listed as the event owner in the managers_and_staff payload.
	// Check separately here instead of modifying buildCreateOrUpdatePayload() for minimal change.
	isOwner, err := s.createEventCheckOwnerIsRequester(req.ManagersAndStaff, userIdUUID, ctx)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("User must list themselves as the event's owner in managers_and_staff")
	}

	payload, err := buildCreateOrUpdatePayload(req)
	if err != nil {
		return nil, err
	}
	return s.repo.Event.CreateEvent(ctx, payload)
}

func (s *service) UpdateEvent(ctx context.Context, id string, userId string, req dtoReq.UpdateEventReq) (*dtoRes.UpdateEventRes, error) {
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("Invalid event id")
	}
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, errors.New("Invalid user id")
	}

	role, err := s.repo.Event.GetUserRoleInEvent(idUUID, userIdUUID, ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if role == nil || (*role != string(entity.OWNER) && *role != string(entity.MANAGER)) {
		return nil, errors.New("Cannot update event; user is not owner or manager")
	}

	payload, err := buildCreateOrUpdatePayload(dtoReq.CreateEventReq(req))
	if err != nil {
		return nil, err
	}
	return s.repo.Event.UpdateEvent(ctx, id, payload)
}

func buildCreateOrUpdatePayload(req dtoReq.CreateEventReq) (entity.CreateEventPayload, error) {
	if err := validateThaiTimezone(req.Timezone); err != nil {
		return entity.CreateEventPayload{}, err
	}

	if req.AllowAllToScan == nil {
		return entity.CreateEventPayload{}, fmt.Errorf("allow_all_to_scan is required")
	}

	at, err := entity.ParseAttendanceType(req.AttendanceType)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	// Time parsing and validation
	startTime, err := parseTime(req.StartTime)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}
	endTime, err := parseTime(req.EndTime)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	if !endTime.After(startTime) {
		return entity.CreateEventPayload{}, fmt.Errorf("end_time must be after start_time")
	}
	if !isSameDay(startTime, endTime) {
		return entity.CreateEventPayload{}, fmt.Errorf("start_time and end_time must be on the same day in timezone %s", entity.ThaiTZ)
	}

	// attendance_type=all -> attendee must be empty
	if at == entity.AttendanceAll && len(req.Attendee) != 0 {
		return entity.CreateEventPayload{}, fmt.Errorf("attendee must be empty when attendance_type=all")
	}

	revealedFields, err := entity.ParseParticipantFields(req.RevealedFields)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	locationPoint := entity.Point{
		X: req.LocationLat,
		Y: req.LocationLong,
	}

	event := entity.Event{
		Name:        req.Name,
		Organizer:   req.Organizer,
		Description: &req.Description,

		StartTime: startTime,
		EndTime:   endTime,

		Location:       req.Location,
		LocationPoint:  locationPoint,
		AttendenceType: at,
		AllowAllToScan: *req.AllowAllToScan,
		EvaluationForm: &req.EvaluationForm,
		RevealedFields: revealedFields,
	}

	agendas, err := buildAgendas(req.Agenda, startTime, endTime)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	whitelist, faculties, err := buildAttendanceTargets(at, req.Attendee)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	// managers_and_staff -> EventUsersInput (ref_id + parsed role)
	eventUsersInput, err := buildEventUsersInput(req.ManagersAndStaff)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	return entity.CreateEventPayload{
		Event:            event,
		Agendas:          agendas,
		Whitelist:        whitelist,
		AllowedFaculties: faculties,
		EventUsersInput:  eventUsersInput,
	}, nil

}

func buildEventUsersInput(in []dtoReq.ManagerStaffReq) ([]entity.EventUserInput, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]entity.EventUserInput, 0, len(in))
	seenRole := make(map[uint64]string, len(in)) // ref_id -> role string
	ownerCount := 0

	for _, m := range in {
		r, err := entity.ParseRole(m.Role)
		if err != nil {
			return nil, err
		}

		rs := string(r)
		if old, ok := seenRole[m.RefID]; ok {
			if old != rs {
				return nil, fmt.Errorf("duplicate ref_id with different role in managers_and_staff: %d", m.RefID)
			}
			continue
		}

		seenRole[m.RefID] = rs
		if r == entity.OWNER {
			ownerCount += 1
		}

		out = append(out, entity.EventUserInput{
			RefID: m.RefID,
			Role:  r,
		})
	}

	if ownerCount != 1 {
		return nil, fmt.Errorf("require exactly one owner in manager_and_staff")
	}

	return out, nil
}

// validateThaiTimezone enforces client timezone to be Asia/Bangkok only.
func validateThaiTimezone(tz string) error {
	if strings.TrimSpace(tz) != entity.ThaiTZ {
		return fmt.Errorf("timezone must be %s", entity.ThaiTZ)
	}
	return nil
}

// parseTime parses RFC3339 and enforces that it is UTC (Z / +00:00).
func parseTime(v string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(v))
	if err != nil {
		return time.Time{}, err
	}
	_, off := t.Zone()
	if off != 0 {
		return time.Time{}, fmt.Errorf("time must be UTC (use Z / +00:00)")
	}
	return t.UTC(), nil
}

// isSameDay checks whether two UTC instants are on the same calendar day in Thai timezone.
func isSameDay(aUTC, bUTC time.Time) bool {
	a := aUTC.In(thaiLoc)
	b := bUTC.In(thaiLoc)
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// isSameDay checks whether two UTC instants are on the same calendar day in Thai timezone.
func buildAgendas(in []dtoReq.CreateAgendaReq, eventStartUTC, eventEndUTC time.Time) ([]entity.EventAgenda, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]entity.EventAgenda, 0, len(in))
	for _, a := range in {
		st, err := parseTime(a.StartTime)
		if err != nil {
			return nil, err
		}
		et, err := parseTime(a.EndTime)
		if err != nil {
			return nil, err
		}

		if !et.After(st) {
			return nil, fmt.Errorf("agenda end_time must be after start_time")
		}
		if !isSameDay(st, et) {
			return nil, fmt.Errorf("agenda start_time and end_time must be on the same day in timezone %s", entity.ThaiTZ)
		}
		if st.Before(eventStartUTC) || et.After(eventEndUTC) {
			return nil, fmt.Errorf("agenda time must be within event start_time and end_time")
		}

		out = append(out, entity.EventAgenda{
			ActivityName: a.ActivityName,
			StartTime:    st,
			EndTime:      et,
		})
	}
	return out, nil
}

// buildAttendanceTargets converts attendee list into whitelist/faculty rows based on attendance_type.
func buildAttendanceTargets(at entity.AttendanceType, attendee []any) ([]entity.EventWhitelist, []entity.EventAllowedFaculties, error) {
	switch at {
	case entity.AttendanceAll:
		return nil, nil, nil

	case entity.AttendanceWhitelist:
		if len(attendee) == 0 {
			return nil, nil, fmt.Errorf("attendee is required")
		}
		out := make([]entity.EventWhitelist, 0, len(attendee))
		for _, v := range attendee {
			ref, err := anyToUint64(v)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, entity.EventWhitelist{AttendeeRefID: ref})
		}
		return out, nil, nil

	case entity.AttendanceFaculties:
		if len(attendee) == 0 {
			return nil, nil, fmt.Errorf("attendee is required")
		}
		out := make([]entity.EventAllowedFaculties, 0, len(attendee))
		for _, v := range attendee {
			fno, err := anyToUint8(v)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, entity.EventAllowedFaculties{FacultyNO: fno})
		}
		return nil, out, nil

	default:
		return nil, nil, fmt.Errorf("invalid attendance_type")
	}
}

// anyToUint64 converts JSON number/string into uint64.
func anyToUint64(v any) (uint64, error) {
	switch x := v.(type) {
	case uint64:
		return x, nil
	case int:
		if x < 0 {
			return 0, fmt.Errorf("invalid number")
		}
		return uint64(x), nil
	case int64:
		if x < 0 {
			return 0, fmt.Errorf("invalid number")
		}
		return uint64(x), nil
	case float64:
		if x < 0 {
			return 0, fmt.Errorf("invalid number")
		}
		return uint64(x), nil
	case string:
		u, err := strconv.ParseUint(strings.TrimSpace(x), 10, 64)
		if err != nil {
			return 0, err
		}
		return u, nil
	default:
		return 0, fmt.Errorf("invalid type: %T", v)
	}
}

// anyToUint8 converts JSON number/string into uint8.
func anyToUint8(v any) (uint8, error) {
	u, err := anyToUint64(v)
	if err != nil {
		return 0, err
	}
	if u > 255 {
		return 0, fmt.Errorf("out of range")
	}
	return uint8(u), nil
}

// POST /events helper function.
// Check if the requester is listed as the owner of event
func (s *service) createEventCheckOwnerIsRequester(req []dtoReq.ManagerStaffReq, userId uuid.UUID, ctx context.Context) (bool, error) {
	user, err := s.repo.Auth.GetUserById(datatypes.UUID(userId), ctx)
	if err != nil {
		return false, err
	}

	for _, person := range req {
		if person.Role == string(entity.OWNER) && user.RefID != person.RefID {
			return false, nil
		}
	}
	return true, nil
}
