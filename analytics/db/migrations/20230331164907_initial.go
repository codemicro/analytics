package migrations

import (
	"context"
	"github.com/codemicro/analytics/analytics/db/models"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

func init() {
	logger := log.With().Str("migration", "20230331164907").Logger()

	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		logger.Info().Msg("up")

		if _, err := db.NewCreateTable().Model(&models.Session{}).Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewCreateTable().Model(&models.Request{}).ForeignKey("(session_id) REFERENCES sessions(id)").Exec(ctx); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		logger.Info().Msg("down")

		if _, err := db.NewDropTable().Model(&models.Request{}).Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewDropTable().Model(&models.Session{}).Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}
