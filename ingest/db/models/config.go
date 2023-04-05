package models

import "github.com/uptrace/bun"

type Config struct {
	bun.BaseModel `bun:"config"`

	ID    string `bun:",pk"`
	Value string
}