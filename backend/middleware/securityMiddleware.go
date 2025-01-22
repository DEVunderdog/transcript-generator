package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SecurityHeaderMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
        ctx.Writer.Header().Set("X-Frame-Options", "DENY")
        ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		ctx.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
        ctx.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        ctx.Writer.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		ctx.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self';")
        ctx.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        ctx.Writer.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
        ctx.Writer.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
        ctx.Writer.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
        ctx.Writer.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		ctx.Writer.Header().Del("Server")
        ctx.Writer.Header().Del("X-Powered-By")

		if ctx.Request.Header.Get("X-Forwarded-Proto") != "https" && gin.Mode() == gin.ReleaseMode {
            ctx.Redirect(http.StatusPermanentRedirect, "https://"+ctx.Request.Host+ctx.Request.RequestURI)
            ctx.Abort()
            return
        }
	}
}