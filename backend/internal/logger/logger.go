package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)


type LoggerConfig struct {
	LogLevel zerolog.Level
}

type Logger struct {
	*zerolog.Logger
	config LoggerConfig
}

func NewLogger(config LoggerConfig) *Logger {
	var writers []io.Writer

	consoleWriter := zerolog.ConsoleWriter{
		Out: os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	writers = append(writers, consoleWriter)


	multiWriter := zerolog.MultiLevelWriter(writers...)
	logger := zerolog.New(multiWriter).
		Level(config.LogLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{
		Logger: &logger,
		config: config,
	}
}
