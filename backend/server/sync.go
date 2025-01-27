package server

import (
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/DEVunderdog/transcript-generator-backend/constants"
	database "github.com/DEVunderdog/transcript-generator-backend/database/sqlc"
	"github.com/DEVunderdog/transcript-generator-backend/token"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

type matchResult struct {
	matchedResults   []int32
	unmatchedResults []int32
}

func (server *Server) sync(ctx *gin.Context) {

	listOfFilesInBucket := make(map[string]struct{})

	payload := ctx.MustGet(constants.PayloadKey).(token.Payload)
	userID := strconv.Itoa(int(payload.UserID))

	conflictingFiles, err := server.store.ListConflictingFiles(
		ctx,
		database.ListConflictingFilesParams{
			UserID:              int32(payload.UserID),
			FirstLockCondition:  database.Locked,
			FirstUploadStatus:   database.Success,
			SecondLockCondition: database.Locked,
			SecondUploadStatus:  database.Failed,
			ThirdLockCondition:  database.Locked,
			ThirdUploadStatus:   database.Pending,
			FourthLockCondition: database.Unlocked,
			FourthUploadStatus:  database.Pending,
			FifthLockCondition:  database.Unlocked,
			FifthUploadStatus:   database.Failed,
		},
	)

	if err != nil {
		server.baseLogger.Error().Err(err).Msg("error getting conflicting files from registry")
		server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error getting conflicted files from registry", nil)
		return
	}

	if len(conflictingFiles) == 0 {
		server.enhanceHTTPResponse(ctx, http.StatusOK, "none conflicting files found", nil)
		return
	}

	it := server.storageClient.StorageClient.Bucket(server.storageClient.BucketName).Objects(ctx, &storage.Query{
		Prefix: userID + "/",
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			server.baseLogger.Error().Err(err).Msg("error while fetching list of object keys from bucket")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while fetching list of conflicting files", nil)
			return
		}

		listOfFilesInBucket[attrs.Name] = struct{}{}
	}

	if len(listOfFilesInBucket) == 0 {
		fileIDs := make([]int32, 0, len(conflictingFiles))
		for _, item := range conflictingFiles {
			fileIDs = append(fileIDs, item.ID)
		}

		err := server.store.DeleteMultipleFilesTx(ctx, int32(payload.UserID), fileIDs)
		if err != nil {
			server.baseLogger.Error().Err(err).Msg("error deleting conflicting files")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error cleaning up files, sync up later", nil)
			return
		}

		server.enhanceHTTPResponse(ctx, http.StatusOK, "cleaned up successfully", nil)
		return
	}

	results := matchResult{
		matchedResults:   make([]int32, 0),
		unmatchedResults: make([]int32, 0),
	}

	for _, item := range conflictingFiles {
		_, exists := listOfFilesInBucket[item.ObjectKey.String]
		if exists {
			results.matchedResults = append(results.matchedResults, item.ID)
		} else {
			results.unmatchedResults = append(results.unmatchedResults, item.ID)
		}
	}

	if len(results.matchedResults) != 0 {
		err := server.store.UnlockMultipleFilesTx(ctx, int32(payload.UserID), results.matchedResults)
		if err != nil {
			server.baseLogger.Error().Err(err).Msg("error while unlocking matched files")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while cleaning up files", nil)
			return
		}
	}

	if len(results.unmatchedResults) != 0 {
		err := server.store.DeleteMultipleFilesTx(ctx, int32(payload.UserID), results.unmatchedResults)
		if err != nil {
			server.baseLogger.Error().Err(err).Msg("error while cleaning up unmatched files")
			server.enhanceHTTPResponse(ctx, http.StatusInternalServerError, "error while cleaning up files", nil)
			return
		}
	}

	server.enhanceHTTPResponse(ctx, http.StatusOK, "successfully sync up files", nil)

}
