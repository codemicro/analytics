package migrations

import (
	"context"
	"github.com/codemicro/analytics/analytics/db/models"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

func init() {
	logger := log.With().Str("migration", "20230331164907").Logger()

	tables := []any{
		&models.Request{},
	}

	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		logger.Info().Msg("up")

		for _, table := range tables {
			if _, err := db.NewCreateTable().Model(table).Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		logger.Info().Msg("down")

		for _, table := range tables {
			if _, err := db.NewDropTable().Model(table).Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
}
