package db

import (
	"context"
	"database/sql"
	"github.com/codemicro/analytics/analytics/config"
	"github.com/codemicro/analytics/analytics/db/migrations"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
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
