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

		if _, err := db.NewCreateTable().Model(&models.Config{}).Exec(ctx); err != nil {
			return err
		}

		if _, err := db.NewInsert().Model(&[]*models.Config{
			{ID: "session_inactive_after", Value: "+2 hours"},
			{ID: "prune_requests_after", Value: "+14 days"},
		}).Exec(ctx); err != nil {
			return err
		}

		if _, err := db.NewCreateTable().Model(&models.Session{}).Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewCreateTable().Model(&models.Request{}).ForeignKey("(session_id) REFERENCES sessions(id)").Exec(ctx); err != nil {
			return err
		}

		views := []string{
			`CREATE VIEW "sessions_with_last_seen" AS SELECT *, (SELECT MAX("time") FROM "requests" WHERE "session_id" = "session"."id") AS "last_seen" FROM "sessions" AS "session";`,
			`CREATE VIEW "active_sessions" AS SELECT * FROM "sessions_with_last_seen" WHERE datetime() < datetime("last_seen", (SELECT "value" FROM "config" WHERE id='session_inactive_after'));`,
			`CREATE VIEW "top_urls" AS SELECT "host", "uri", COUNT(*) as "count" FROM "requests" GROUP BY "host", "uri" ORDER BY "count" DESC;`,
			`CREATE VIEW "top_user_agents" AS SELECT "user_agent", COUNT(*) AS "count" FROM "requests" GROUP BY "user_agent" ORDER BY "count" DESC;`,
			`CREATE VIEW "top_ip_addresses" AS SELECT "ip_addr", COUNT(*) AS "count" FROM "requests" GROUP BY "id_addr" ORDER BY "count" DESC;`,
			`CREATE VIEW "top_referers" AS SELECT "referer", COUNT(*) AS "count" FROM "requests" GROUP BY "referer" ORDER BY "count" DESC;`,
		}

		for _, query := range views {
			if _, err := db.Exec(query); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		logger.Info().Msg("down")

		if _, err := db.NewDropTable().Model(&models.Config{}).Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewDropTable().Model(&models.Request{}).Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewDropTable().Model(&models.Session{}).Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}
