package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type healthUpdates struct {
	ServerStatus string `json:"server_status"`
}

func (server *Server) serverHealthCheck(ctx *gin.Context) {
	healthUpdates := healthUpdates {
		ServerStatus: "up and running",
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "server status", healthUpdates)
}
