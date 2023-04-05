package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/codemicro/analytics/ingest/config"
	"github.com/codemicro/analytics/ingest/db/migrations"
	"github.com/codemicro/analytics/ingest/db/models"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
)

type DB struct {
	DB *bun.DB
}

func New(conf *config.Config) (*DB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, conf.Database.DSN)
	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())
	if config.Debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	log.Info().Msg("migrating database")
	mig := migrate.NewMigrator(db, migrations.Migrations)
	if err := mig.Init(context.Background()); err != nil {
		return nil, err
	}
	if group, err := mig.Migrate(context.Background()); err != nil {
		return nil, err
	} else if group.IsZero() {
		log.Info().Msg("no migrations to run (database is up-to-date)")
	} else {
		log.Info().Msg("migrations completed")
	}

	return &DB{
		DB: db,
	}, nil
}

func (db *DB) GetSessionsWithActivityAfter(minutes int, sort string) ([]*models.Session, error) {
	var sessions []*models.Session
	q := db.DB.NewSelect().
		Model((*models.Session)(nil)).
		ColumnExpr("*").
		ColumnExpr(`(select max("time") as "time" from requests where session_id = "session"."id") as "last_seen"`)
	if sort != "" {
		q = q.Order(sort)
	}
	if minutes > 0 {
		q = q.Where(fmt.Sprintf(`datetime() < datetime("last_seen", '+%d minutes')`, minutes))
	}
	if err := q.Scan(context.Background(), &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}
