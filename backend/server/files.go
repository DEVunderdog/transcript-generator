package server

import (
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/DEVunderdog/transcript-generator-backend/constants"
	database "github.com/DEVunderdog/transcript-generator-backend/database/sqlc"
	custom_errors "github.com/DEVunderdog/transcript-generator-backend/errors"
	"github.com/DEVunderdog/transcript-generator-backend/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const maxFileSize = 50 * 1024 * 1024

type uploadFileRequest struct {
	Filename string `json:"filename" binding:"required"`
	Filepath string `json:"filepath" binding:"required"`
}

type uploadedFileResponse struct {
	ID       int32  `json:"id"`
	FileName string `json:"filenmae"`
}

func (server *Server) uploadFileToBucket(ctx *gin.Context) {

	var req uploadFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "please provide valid request body", err.Error())
		return
	}

	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)

	newFile, err := server.store.CreateEmptyFileTx(ctx, database.CreateEmptyFileParams{
		UserID:   int32(payload.UserID),
		FileName: req.Filename,
	})

	if err != nil {
		if errors.Is(err, custom_errors.ErrDuplicateData) {
			server.baseLogger.Logger.Error().Err(err).Msg("file with that name already exists")
			server.enhanceHTTPResponse(ctx, http.StatusConflict, "file with that name already exits", nil)
			return
		}

		server.baseLogger.Logger.Error().Err(err).Msg("error while creating empty file in registry")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating file in registry", nil)
		return
	}

	fileInfo, err := os.Stat(req.Filepath)
	if err != nil {
		if os.IsNotExist(err) {
			server.baseLogger.Logger.Error().Err(err).Msg("file not found on requested path")
			server.enhanceHTTPResponse(ctx, http.StatusNotFound, "file not found on requested path", nil)
			return
		}
		server.baseLogger.Logger.Error().Err(err).Msg("error while reading file from provided path")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while reading file from provided path", nil)
		return
	}

	if fileInfo.Size() > maxFileSize {
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "file size exceeded then 50 mb", nil)
		return
	}

	f, err := os.Open(req.Filepath)
	if err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while opening local file for uploading", err.Error())
		return
	}
	defer f.Close()

	objectName := uuid.NewString() + "-" + req.Filename
	object := server.storageClient.StorageClient.Bucket(server.storageClient.BucketName).Object(objectName)

	writer := object.NewWriter(ctx)

	if _, err := io.Copy(writer, f); err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "failed to copy file to storage", err.Error())
		return
	}

	if err := writer.Close(); err != nil {
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "failed to close writer which is used to writing storage", err.Error())
		return
	}

	file, err := server.store.UpdateMetadataFileTx(ctx, database.UpdateFileMetadataTxParams{
		ID: newFile.ID,
		ObjectKey: pgtype.Text{
			Valid:  true,
			String: objectName,
		},
		UpdatedAt: newFile.UpdatedAt,
	})

	if err != nil {
		if errors.Is(err, custom_errors.ErrNoRecordFound) {
			server.baseLogger.Logger.Error().Err(err).Msg("cannot find the empty file which was created earlier")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating file in registry, please try later", nil)
			return
		}

		if errors.Is(err, custom_errors.ErrResourceConflict) {
			server.baseLogger.Logger.Error().Err(err).Msg("resource concurrently got tampered")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "resource conflicts, please try again later or maybe sync up", nil)
			return
		}

		server.baseLogger.Logger.Error().Err(err).Msg("error while updating metadata of empty file")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating file in registry", nil)
		return

	}

	server.enhanceHTTPResponse(ctx, http.StatusCreated, "file uploaded successfully", uploadedFileResponse{
		ID:       file.ID,
		FileName: file.FileName,
	})
}
