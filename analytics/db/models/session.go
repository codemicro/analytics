package models

import (
	"github.com/uptrace/bun"
)

type Session struct {
	bun.BaseModel

	ID        string `bun:",pk"`
	UserAgent string `bun:"type:VARCHAR COLLATE NOCASE"`
	IPAddr    string
}
