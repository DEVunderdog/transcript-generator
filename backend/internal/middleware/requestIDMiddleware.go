package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDKey = "RequestID"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader("X-Request-ID")

		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx.Set(requestIDKey, requestID)
		ctx.Header("X-Request-ID", requestID)

		ctx.Next()
	}
}