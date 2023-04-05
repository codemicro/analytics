//go:build debug

package debug

import (
	"fmt"
	"github.com/rs/zerolog"
)

var Enable = true

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	fmt.Println("DEBUG MODE ACTIVE")
	fmt.Println()
}
