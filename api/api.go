// api.go

// Package api must be documented
package api

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	apiLogger zerolog.Logger
)

// InitLogger initialize the sublogger for API
func InitLogger() {
	apiLogger = log.With().Str("module", "API").Logger()
}
