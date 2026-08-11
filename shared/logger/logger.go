// Package logger provides the zerolog logger setup shared by RVPay
// services: a timestamped, caller-annotated logger at a validated level.
//
// Responsibility: construct a zerolog.Logger with a consistent format and
// an explicit level.
//
// Consumers: service bootstrap code (cmd/grpc-service).
//
// Non-responsibilities: it does not configure output sinks, emit logs, or
// know about service-specific configuration.
package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

// New returns a zerolog.Logger writing to w with timestamps and caller
// information at the given level. An empty level defaults to "info". A nil
// writer defaults to os.Stderr. The level must be parseable by
// zerolog.ParseLevel.
func New(level string, w io.Writer) (zerolog.Logger, error) {
	if w == nil {
		w = os.Stderr
	}
	if level == "" {
		level = "info"
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("parse log level %q: %w", level, err)
	}

	return zerolog.New(w).With().Timestamp().Caller().Logger().Level(lvl), nil
}