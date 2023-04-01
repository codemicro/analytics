package config

import "github.com/codemicro/analytics/analytics/config/internal/debug"

var Debug = debug.Enable

type Config struct {
	Ingest struct {
		Address string
	}
	Database struct {
		DSN string
	}
	HTTP struct {
		Address string
	}
}

func Load() (*Config, error) {
	cl := new(configLoader)
	if err := cl.load("config.yml"); err != nil {
		return nil, err
	}

	conf := new(Config)
	conf.Ingest.Address = asString(cl.withDefault("ingest.address", "127.0.0.1:7500"))
	conf.HTTP.Address = asString(cl.withDefault("http.address", "127.0.0.1:8080"))
	conf.Database.DSN = asString(cl.withDefault("database.dsn", "analytics.db"))

	return conf, nil
}
