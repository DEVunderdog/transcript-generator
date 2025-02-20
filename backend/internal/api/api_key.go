package api

import (
	"net/http"

	"github.com/DEVunderdog/transcript-generator-backend/internal/constants"
	database "github.com/DEVunderdog/transcript-generator-backend/internal/database/sqlc"
	"github.com/DEVunderdog/transcript-generator-backend/internal/token"
	"github.com/gin-gonic/gin"
)

type apiKeyRequest struct {
	Email string `json:"email" binding:"required"`
}

type apiKeyResponse struct {
	APIKey string `json:"api_key"`
}

// @Summary Generate API Key
// @Description Registers a user and generates an API Key
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body apiKeyRequest true "User Email"
// @Success 201 {object} apiKeyResponse "api keys created"
// @Failure 400 {object} standardResponse "invalid request"
// @Failure 403 {object} standardResponse "user already present"
// @Failure 500 {object} standardResponse "Internal Server Error"
// @Router /api/register [POST]
func (server *Server) generateAPIKey(ctx *gin.Context) {

	var req apiKeyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "please provide valid request", err.Error())
		return
	}

	user, err := server.store.CreateUsers(ctx, req.Email)
	if err != nil {
		if database.ErrorCode(err) == database.UniqueViolation {
			server.enhanceHTTPResponse(ctx, http.StatusForbidden, "user already present", err.Error())
			return
		}
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating user in database", err.Error())
		return
	}

	apiKey, signature, err := token.GenerateAndSignAPIKey(server.tokenMaker.PrivateKey)
	if err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error creating api keys", err.Error())
		return
	}

	_, err = server.store.CreateAPIKey(ctx, database.CreateAPIKeyParams{
		UserID:     user.ID,
		Credential: []byte(apiKey),
		Signature:  signature,
	})

	if err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating api keys in database", err.Error())
		return
	}

	server.enhanceHTTPResponse(ctx, http.StatusCreated, "api keys created", apiKeyResponse{
		APIKey: apiKey,
	})
}

// @Summary Delete API Key
// @Description Request to delete the API Key
// @Tags Authentication
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} standardResponse "api key deleted successfully"
// @Failure 500 {object} standardResponse "Internal Server Error"
// @Router /auth/api/delete [DELETE]
func (server *Server) deleteAPIKey(ctx *gin.Context) {

	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)

	err := server.store.DeleteAPIKey(ctx, []byte(payload.APIKey))
	if err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while deleting api keys from database", err.Error())
		return
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "api key deleted successfully", nil)
}
