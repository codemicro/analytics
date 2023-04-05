package config

import (
	"github.com/codemicro/analytics/analytics/config/internal/debug"
)

var Debug = debug.Enable

type Config struct {
	Ingest struct {
		Address string
	}
	Database struct {
		DSN string
	}
}

func Load() (*Config, error) {
	cl := new(configLoader)
	if err := cl.load("config.yml"); err != nil {
		return nil, err
	}

	conf := new(Config)
	conf.Ingest.Address = asString(cl.withDefault("ingest.address", "127.0.0.1:7500"))
	conf.Database.DSN = asString(cl.withDefault("database.dsn", "analytics.db"))

	return conf, nil
}
