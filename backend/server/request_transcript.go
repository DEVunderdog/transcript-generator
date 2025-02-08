package server

import (
	"errors"
	"net/http"

	"github.com/DEVunderdog/transcript-generator-backend/constants"
	database "github.com/DEVunderdog/transcript-generator-backend/database/sqlc"
	"github.com/DEVunderdog/transcript-generator-backend/token"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type transcriptRequestQuery struct {
	Filename string `form:"filename" binding:"required"`
}

func (server *Server) requestTranscript(ctx *gin.Context) {
	var query transcriptRequestQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		server.baseLogger.Error().Err(err).Msg("bad request body")
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "bad request body", err.Error())
		return
	}

	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)

	file, err := server.store.GetFileByName(ctx, database.GetFileByNameParams{
		FileName: query.Filename,
		UserID:   int32(payload.UserID),
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			server.baseLogger.Error().Err(err).Msg("cannot find record")
			server.enhanceHTTPResponse(ctx, http.StatusNotFound, "cannot find the file with provided name", nil)
			return
		}

		server.baseLogger.Error().Err(err).Msg("error while retrieving file details for requesting transcript")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error retrieveing file details for transcript", nil)
		return
	}

	err = server.pubsubClient.PublishMessage(ctx, file.ObjectKey.String, payload.UserID)
	if err != nil {
		server.baseLogger.Error().Err(err).Msg("error publishing message to topic")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error requesting for transcript to service", nil)
		return
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "transcript requested successfully", nil)

}
