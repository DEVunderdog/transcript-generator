package middleware

import (
	"time"

	"github.com/DEVunderdog/transcript-generator-backend/logger"
	"github.com/gin-gonic/gin"
)

type HTTPLogger struct {
	*logger.Logger
}

func NewHTTPLogger(baseLogger *logger.Logger) *HTTPLogger {
	return &HTTPLogger{
		Logger: baseLogger,
	}
}

func (l *HTTPLogger) LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		if raw != "" {
			path = path + "?" + raw
		}

		requestID, exists := ctx.Get(requestIDKey)
		if !exists {
			requestID = "no-request-id"
		}

		ctx.Next()

		event := l.Info().
			Str("service", "http").
			Str("method", ctx.Request.Method).
			Int("status", ctx.Writer.Status()).
			Str("path", path).
			Dur("latency", time.Since(start)).
			Str("client_ip", ctx.ClientIP()).
			Int("body_size", ctx.Writer.Size()).
			Any("request_id", requestID)

		if len(ctx.Errors) > 0 {
			event.Strs("errors", ctx.Errors.Errors())
		}

		event.Msg("http_request")
	}
}
