package ingest

import (
	"bufio"
	"errors"
	"github.com/codemicro/analytics/analytics/config"
	"github.com/codemicro/analytics/analytics/db"
	"github.com/rs/zerolog/log"
	"io"
	"net"
)

type Ingest struct {
	db       *db.DB
	listener net.Listener
}

func Start(conf *config.Config, database *db.DB) (*Ingest, error) {
	ingest := &Ingest{
		db: database,
	}

	var err error
	ingest.listener, err = net.Listen("tcp", conf.Ingest.Address)
	if err != nil {
		return nil, err
	}

	go ingest.serveConnections()

	log.Info().Msgf("listener alive on %s", ingest.listener.Addr().String())

	return ingest, nil
}

func (i *Ingest) Stop() error {
	return i.listener.Close()
}

func (i *Ingest) serveConnections() {
	for {
		conn, err := i.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Error().Err(err).Msg("unhandled error when accepting ingest connection")
			continue
		}
		go i.processConnection(conn)
	}
}

func (i *Ingest) processConnection(conn net.Conn) {
	defer conn.Close()

	log.Debug().Str("remote_address", conn.RemoteAddr().String()).Msg("new connection")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		i.processLog([]byte(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		if !errors.Is(err, io.EOF) {
			log.Error().Err(err).Msg("unable to scan from connection")
			return
		}
	}

	log.Debug().Str("remote_address", conn.RemoteAddr().String()).Msg("closing connection")
}
