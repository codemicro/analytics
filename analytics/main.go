package main

import (
	"github.com/codemicro/analytics/analytics/config"
	"github.com/codemicro/analytics/analytics/db"
	"github.com/codemicro/analytics/analytics/ingest"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("unhandled error")
	}
}

func run() error {
	conf, err := config.Load()
	if err != nil {
		return err
	}

	database, err := db.New(conf)
	if err != nil {
		return err
	}

	ig, err := ingest.Start(conf, database)
	if err != nil {
		return err
	}

	waitForSignal(syscall.SIGINT)

	log.Info().Msg("terminating")

	_ = ig.Stop()
	return nil
}

func waitForSignal(sig syscall.Signal) {
	cchan := make(chan os.Signal)
	signal.Notify(cchan, sig)
	<-cchan
}
