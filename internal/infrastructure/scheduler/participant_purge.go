package scheduler

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/repository"
	"github.com/rs/zerolog"
)


func RunParticipantPurge(ctx context.Context, repo repository.AllRepo, retentionDays int, logger *zerolog.Logger) {
	purged, err := repo.Event.PurgeStaleParticipants(ctx, retentionDays)
	if err != nil {
		logger.Error().Err(err).Int64("purged", purged).Msg("participant purge finished with errors")
		return
	}
	logger.Info().Int64("purged", purged).Msg("participant purge finished")
}
