//go:build !debug

package debug

import (
	"github.com/rs/zerolog"
)

var Enable = false

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
