package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/codemicro/analytics/analytics/db/models"
	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog/log"
	"math"
	"net/url"
	"time"
)

func (i *Ingest) processLog(inp []byte) {
	cl := new(CaddyLog)
	if err := json.Unmarshal(inp, cl); err != nil {
		log.Warn().Err(err).Bytes("raw_input", inp).Msg("remote sending invalid JSON")
		return
	}

	log.Debug().Msgf("got log on path %s", cl.Request.URI)

	req, err := cl.ToRequestModel()
	if err != nil {
		log.Error().Err(err).Bytes("raw_json", inp).Msg("could not convert CaddyLog to Request")
		return
	}

	tx, err := i.db.DB.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		log.Error().Err(err).Msg("failed to start transaction")
		return
	}

	sess, err := i.assignToSession(tx, req)
	if err != nil {
		log.Error().Err(err).Msg("failed to assign session to request")
	}

	req.SessionID = sess.ID

	if _, err := tx.NewInsert().Model(req).Exec(context.Background()); err != nil {
		log.Error().Err(err).Msg("could not save request into database")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("unable to commit transaction")
		return
	}
}

type CaddyLog struct {
	Level     string  `json:"level"`
	Timestamp float64 `json:"ts"`
	Logger    string  `json:"logger"`
	Message   string  `json:"msg"`
	Request   struct {
		RemoteIP   string              `json:"remote_ip"`
		RemotePort string              `json:"remote_port"`
		Protocol   string              `json:"proto"`
		Method     string              `json:"method"`
		Host       string              `json:"host"`
		URI        string              `json:"uri"`
		Headers    map[string][]string `json:"headers"`
		TLS        struct {
			Resumed     bool   `json:"resumed"`
			Version     int    `json:"version"`
			CipherSuite int    `json:"cipher_suite"`
			Proto       string `json:"proto"`
			ServerName  string `json:"server_name"`
		} `json:"tls"`
	} `json:"request"`
	Duration        float64             `json:"duration"`
	Size            int                 `json:"size"`
	Status          int                 `json:"status"`
	ResponseHeaders map[string][]string `json:"resp_headers"`
}

func (cl *CaddyLog) getRequestHeader(key string) string {
	v, found := cl.Request.Headers[key]
	if !found {
		return ""
	}
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func (cl *CaddyLog) ToRequestModel() (*models.Request, error) {
	parsedURL, err := url.ParseRequestURI(cl.Request.URI)
	if err != nil {
		return nil, err
	}

	var requestTime time.Time
	{
		s, fs := math.Modf(cl.Timestamp)
		requestTime = time.Unix(int64(s), int64(fs))
	}

	return &models.Request{
		ID:         shortuuid.New(),
		Time:       requestTime,
		IPAddr:     cl.Request.RemoteIP,
		Host:       cl.Request.Host,
		RawURI:     cl.Request.URI,
		URI:        parsedURL.Path,
		Referer:    cl.getRequestHeader("Referer"),
		UserAgent:  cl.getRequestHeader("User-Agent"),
		StatusCode: cl.Status,
	}, nil
}
