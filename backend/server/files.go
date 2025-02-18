package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/DEVunderdog/transcript-generator-backend/constants"
	database "github.com/DEVunderdog/transcript-generator-backend/database/sqlc"
	custom_errors "github.com/DEVunderdog/transcript-generator-backend/errors"
	"github.com/DEVunderdog/transcript-generator-backend/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const maxFileSize = 50 * 1024 * 1024

type uploadedFileResponse struct {
	ID       int32  `json:"id" binding:"true"`
	FileName string `json:"filenmae" binding:"true"`
}

type updateFileRequest struct {
	NewFileName string `json:"new_file_name" binding:"required"`
	FileID      int32  `json:"file_id" binding:"required"`
}

// @Summary Upload file to bucket
// @Description Upload an audio file to the cloud storage.
// @Tags Files
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} standardResponse "File uploaded successfully"
// @Failure 400 {object} standardResponse "Bad Request"
// @Failure 409 {object} standardResponse "Status Conflict"
// @Failure 500 {object} standardResponse "Internal Server Error"
// @Router /auth/files/upload [POST]
func (server *Server) uploadFileToBucket(ctx *gin.Context) {

	file, err := ctx.FormFile("file")
	if err != nil {
		server.baseLogger.Error().Err(err).Msg("no file received")
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "No file received", err.Error())
		return
	}

	if file.Size > maxFileSize {
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "file size exceeded than 50 mb", nil)
		return
	}

	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)

	extension := filepath.Ext(file.Filename)

	newFileName := uuid.New().String() + extension
	objectKey := fmt.Sprintf("%d/%s", payload.UserID, newFileName)

	src, err := file.Open()
	if err != nil {
		server.baseLogger.Error().Err(err).Msg("error reading uploaded file")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error opening uploaded file", nil)
		return
	}
	defer src.Close()

	newFile, err := server.store.CreateEmptyFileTx(ctx, database.CreateEmptyFileParams{
		UserID:       int32(payload.UserID),
		FileName:     file.Filename,
		LockStatus:   database.Locked,
		UploadStatus: database.Pending,
	})

	if err != nil {
		if errors.Is(err, custom_errors.ErrDuplicateData) {
			server.baseLogger.Error().Err(err).Msg("file with that name already exists")
			server.enhanceHTTPResponse(ctx, http.StatusConflict, "file with that name already exits", nil)
			return
		}

		server.baseLogger.Error().Err(err).Msg("error while creating empty file in registry")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating file in registry", nil)
		return
	}

	object := server.storageClient.StorageClient.Bucket(server.storageClient.BucketName).Object(objectKey)

	writer := object.NewWriter(ctx)
	if _, err := io.Copy(writer, src); err != nil {
		rollbackErr := server.store.DeleteFileTx(ctx, int32(payload.UserID), newFile.ID, newFile.UpdatedAt)
		if rollbackErr != nil {
			server.baseLogger.Error().Err(rollbackErr).Msgf("error while rollbacking file by deleting it because writer got failed: %s", err.Error())
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while uploading the file, try later and sync up", nil)
			return
		}
		server.baseLogger.Error().Err(err).Msg("error while copying source of file to cloud storage bucket writer")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while copying source of file to cloud storage bucket writer", nil)
		return
	}

	if err := writer.Close(); err != nil {
		_, rollbackErr := server.store.UpdateMetadataFileTx(ctx, database.UpdateFileMetadataTxParams{
			ID:     newFile.ID,
			UserID: int32(payload.UserID),
			ObjectKey: pgtype.Text{
				Valid:  true,
				String: objectKey,
			},
			UpdatedAt:  newFile.UpdatedAt,
			FileStatus: database.Failed,
		})
		if rollbackErr != nil {
			server.baseLogger.Error().Err(rollbackErr).Msgf("error while rollbacking file by updating its status to failed because writer got error whle closing: %s", err.Error())
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while uploading the file, try later and sync up", nil)
			return
		}
		server.baseLogger.Error().Err(err).Msg("error while closing the cloud storage bucket writer")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while closing cloud storage bucket writer", nil)
		return
	}

	updatedFile, err := server.store.UpdateMetadataFileTx(ctx, database.UpdateFileMetadataTxParams{
		ID: newFile.ID,
		ObjectKey: pgtype.Text{
			Valid:  true,
			String: objectKey,
		},
		UpdatedAt:  newFile.UpdatedAt,
		UserID:     int32(payload.UserID),
		FileStatus: database.Success,
	})

	if err != nil {
		if errors.Is(err, custom_errors.ErrNoRecordFound) {
			server.baseLogger.Error().Err(err).Msg("cannot find the empty file which was created earlier")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating file in registry, please try later", nil)
			return
		}

		if errors.Is(err, custom_errors.ErrResourceConflict) {
			server.baseLogger.Error().Err(err).Msg("resource concurrently got tampered")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "resource conflicts, please try again later or maybe sync up", nil)
			return
		}

		server.baseLogger.Error().Err(err).Msg("error while updating metadata of empty file")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while creating file in registry, sync up or try later", nil)
		return

	}

	server.enhanceHTTPResponse(ctx, http.StatusCreated, "file uploaded successfully", uploadedFileResponse{
		ID:       updatedFile.ID,
		FileName: updatedFile.FileName,
	})
}

// @Summary List Files
// @Description List all files
// @Tags Files
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} standardResponse "file fetched successfully"
// @Failure 500 {object} standardResponse "Internal Server Error"
// @Router /auth/files/list [GET]
func (server *Server) listAllFiles(ctx *gin.Context) {
	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)

	files, err := server.store.ListAllFiles(ctx, database.ListAllFilesParams{
		UserID:       int32(payload.UserID),
		UploadStatus: database.Success,
		LockStatus:   database.Unlocked,
	})

	if err != nil {
		server.baseLogger.Error().Err(err).Msg("error while listing all files")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while listing all files", nil)
		return
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "files fetched successfully", files)
}

// @Summary Update File
// @Description Update file name
// @Tags Files
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body updateFileRequest true "File Name"
// @Success 200 {object} standardResponse "File updated successfully"
// @Failure 400 {object} standardResponse "Bad Request"
// @Failure 404 {object} standardResponse "Not Found"
// @Failure 500 {object} standardResponse "Internal Server Error"
// @Router /auth/files/update [POST]
func (server *Server) updateFile(ctx *gin.Context) {
	var req updateFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.baseLogger.Error().Err(err).Msg("bad request body")
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "bad request body", err.Error())
		return
	}

	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)

	file, err := server.store.UpdateFileName(ctx, database.UpdateFileNameParams{
		NewFileName: req.NewFileName,
		ID:          req.FileID,
		UserID:      int32(payload.UserID),
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			server.baseLogger.Error().Err(err).Msg("cannot find rows")
			server.enhanceHTTPResponse(ctx, http.StatusNotFound, "cannot find the files with provided name", nil)
			return
		}

		server.baseLogger.Error().Err(err).Msg("error while updating files")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while updating file", nil)
		return
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "file updated successfully", uploadedFileResponse{
		ID:       file.ID,
		FileName: file.FileName,
	})
}

// @Summary Delete a file
// @Description Deletes a file from storage
// @Tags Files
// @Security ApiKeyAuth
// @Produce json
// @Param filename path string true "Filename to delete"
// @Success 200 {object} standardResponse "File deleted successfully"
// @Failure 400 {object} standardResponse "Bad Request"
// @Failure 404 {object} standardResponse "File not found"
// @Failure 500 {object} standardResponse "Internal Server Error"
// @Router /auth/files/delete/{filename} [DELETE]
func (server *Server) deleteFile(ctx *gin.Context) {
	fileName := ctx.Param("filename")
	if fileName == "" {
		server.enhanceHTTPResponse(ctx, http.StatusBadRequest, "please provide file name to delete", nil)
		return
	}

	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)

	lockFile, err := server.store.LockFileTx(ctx, int32(payload.UserID), fileName)

	if err != nil {
		if errors.Is(err, custom_errors.ErrNoRecordFound) {
			server.baseLogger.Error().Err(err).Msg("cannot find the file with provide name")
			server.enhanceHTTPResponse(ctx, http.StatusNotFound, "cannot find file with provided name", nil)
			return
		}

		if errors.Is(err, custom_errors.ErrResourceLocked) {
			server.baseLogger.Error().Err(err).Msg("resource locked")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "resource locked, maybe try later", nil)
			return
		}

		server.baseLogger.Error().Err(err).Msg("error while locking the file")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while deleting the file, please sync up or try later", nil)
		return
	}

	object := server.storageClient.StorageClient.Bucket(server.storageClient.BucketName).Object(lockFile.ObjectKey.String)
	if err := object.Delete(ctx); err != nil {
		_, rollbackErr := server.store.UnlockAndLockFile(ctx, database.UnlockAndLockFileParams{
			ID:         lockFile.ID,
			UserID:     int32(payload.UserID),
			LockStatus: database.Unlocked,
			Status:     database.Success,
		})

		if rollbackErr != nil {
			server.baseLogger.Error().Err(rollbackErr).Msgf("error while rollbacking by unlocking the file due failed deletion in cloud storage bucked: %s", err.Error())
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while deleting file, please try later", nil)
			return
		}

		server.baseLogger.Error().Err(err).Msg("error while deleting object from bucket")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while deleting file, please try again", nil)
		return
	}

	err = server.store.DeleteFileTx(ctx, int32(payload.UserID), lockFile.ID, lockFile.UpdatedAt)
	if err != nil {
		if errors.Is(err, custom_errors.ErrResourceConflict) {
			server.baseLogger.Error().Err(err).Msg("conflicting resource")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "conflicting resource, please try later or sync up", nil)
			return
		}

		server.baseLogger.Error().Err(err).Msg("error while deleting files")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while deleting file please try later", nil)
		return
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "file deleted successfully", nil)
}
