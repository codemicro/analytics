package models

import (
	"github.com/uptrace/bun"
	"time"
)

type Request struct {
	bun.BaseModel

	ID        string `bun:",pk"`
	Time      time.Time
	IPAddr    string
	Host      string
	RawURI    string
	URI       string
	Referer   string
	UserAgent string

	Session   *Session `bun:"rel:belongs-to,join:session_id=id"`
	SessionID string   `bun:",nullzero"`
}
