package models

import (
	"fmt"
	"github.com/uptrace/bun"
	"time"
)

type Session struct {
	bun.BaseModel

	ID        string `bun:",pk"`
	UserAgent string `bun:"type:VARCHAR COLLATE NOCASE"`
	IPAddr    string

	LastSeen time.Time `bun:",scanonly"`
}

func (s *Session) String() string {
	return fmt.Sprintf("ID:%s UA:%#v IP:%s LastSeen:%s", s.ID, s.UserAgent, s.IPAddr, s.LastSeen.Format(time.DateTime))
}
