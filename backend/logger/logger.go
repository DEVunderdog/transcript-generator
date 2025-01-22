package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type FileConfig struct {
	Path       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

type LoggerConfig struct {
	LogLevel zerolog.Level
	FileConfig *FileConfig
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

	if config.FileConfig != nil {
		fileWriter := &lumberjack.Logger{
			Filename: config.FileConfig.Path,
			MaxSize: config.FileConfig.MaxSize,
			MaxBackups: config.FileConfig.MaxBackups,
			MaxAge: config.FileConfig.MaxAge,
			Compress: config.FileConfig.Compress,
		}
		writers = append(writers, fileWriter)
	}

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
