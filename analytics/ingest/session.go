package ingest

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/analytics/analytics/db/models"
	"github.com/lithammer/shortuuid/v4"
	"github.com/uptrace/bun"
)

func (i *Ingest) assignToSession(tx bun.Tx, request *models.Request) (*models.Session, error) {
	sess := new(models.Session)

	err := tx.NewSelect().
		Model(sess).
		Where("ip_addr = ?", request.IPAddr).
		Where("user_agent = ?", request.UserAgent).
		Where(`? < datetime((select max("time") as "time" from requests where session_id = "session"."id"), '+30 minutes')`, request.Time).
		Scan(context.Background(), sess)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		return sess, nil
	}

	// No session found that matches, create a new one
	sess.ID = shortuuid.New()
	sess.IPAddr = request.IPAddr
	sess.UserAgent = request.UserAgent

	_, err = tx.NewInsert().Model(sess).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return sess, nil
}
