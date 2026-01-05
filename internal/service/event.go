package service

import (
	"context"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type EventService interface {
	GetParticipantService(code string, ctx context.Context) (*[]dtoRes.GetParticipantRes, *response.APIError)
}

func (s *service) GetParticipantService(code string, ctx context.Context) (*[]dtoRes.GetParticipantRes, *response.APIError) {
	return nil, nil
}
