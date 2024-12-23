package util

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).Level(zerolog.InfoLevel).With().Timestamp().Caller().Logger()
