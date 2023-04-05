package worker

import (
	"context"
	"github.com/codemicro/analytics/ingest/db"
	"github.com/codemicro/analytics/ingest/db/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

const interval = time.Hour

func Start(db *db.DB) {
	ticker := time.NewTicker(interval)
	logger := log.Logger.With().Str("location", "worker").Logger()
	run(db, logger)
	go func() {
		for {
			<-ticker.C
			run(db, logger)
		}
	}()
}

func run(db *db.DB, logger zerolog.Logger) {
	logger.Info().Msg("running worker")

	tx, err := db.DB.Begin()
	if err != nil {
		logger.Err(err).Msg("unable to open session")
		_ = tx.Rollback()
		return
	}

	_, err = tx.NewDelete().Model(&models.Request{}).
		Where(`datetime() > datetime("time", (select value from config where id='prune_requests_after'))`).
		Exec(context.Background())
	if err != nil {
		logger.Err(err).Msg("failed to run request delete query")
		_ = tx.Rollback()
		return
	}

	_, err = tx.NewDelete().Model(&models.Session{}).
		Where(`(SELECT COUNT(*) FROM requests WHERE "session_id" = "session"."id") = 0`).
		Exec(context.Background())
	if err != nil {
		logger.Err(err).Msg("failed to run session delete query")
		_ = tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Err(err).Msg("failed to commit transaction")
		_ = tx.Rollback()
		return
	}
}
