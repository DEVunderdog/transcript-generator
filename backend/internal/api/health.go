package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type healthUpdates struct {
	ServerStatus string `json:"server_status"`
}

// @Summary Health Check
// @Description server health check
// @Tags Health
// @Produce json
// @Success 200 {object} standardResponse "server status"
// @Router /health [GET]
func (server *Server) serverHealthCheck(ctx *gin.Context) {
	healthUpdates := healthUpdates{
		ServerStatus: "up and running",
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "server status", healthUpdates)
}
