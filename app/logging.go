package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

// see https://github.com/rs/zerolog?tab=readme-ov-file#pretty-logging
func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05.000"}
	output.FormatLevel = func(i interface{}) string {
		level := strings.ToUpper(fmt.Sprintf("%-6s", i))
		switch level {
		case "DEBUG ":
			return fmt.Sprintf("\033[36m%s\033[0m", level) // Cyan
		case "INFO  ":
			return fmt.Sprintf("\033[32m%s\033[0m", level) // Green
		case "WARN  ":
			return fmt.Sprintf("\033[33m%s\033[0m", level) // Yellow
		case "ERROR ":
			return fmt.Sprintf("\033[31m%s\033[0m", level) // Red
		case "FATAL ":
			return fmt.Sprintf("\033[35m%s\033[0m", level) // Magenta
		default:
			return level
		}
	}
	output.FormatMessage = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("%v", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("\033[2m%s:\033[0m", i) // Dim
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("\033[2m%s\033[0m", i) // Dim
	}

	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixNano
	logger = zerolog.New(output).With().Timestamp().Logger()
}
