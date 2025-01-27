package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/DEVunderdog/transcript-generator-backend/constants"
	database "github.com/DEVunderdog/transcript-generator-backend/database/sqlc"
	"github.com/DEVunderdog/transcript-generator-backend/token"
	"github.com/DEVunderdog/transcript-generator-backend/utils"
	"github.com/gin-gonic/gin"
)

func Authenticate(config utils.Config, store database.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		key, err := token.GetKeyBasedOnPurpose(ctx, store, config.KeysPurpose)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error getting active jwt keys": err.Error()})
			ctx.Abort()
			return
		}

		publicKey, err := token.GetPublicKey([]byte(key.PublicKey))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error getting public key: ": err.Error()})
			ctx.Abort()
			return
		}

		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "please provide token for authorization"})
			ctx.Abort()
			return
		}

		apiString := strings.TrimPrefix(authHeader, "Bearer ")

		parts := strings.SplitN(apiString, " ", 2)
		if len(parts) != 2 || parts[0] != "ApiKey" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			ctx.Abort()
			return
		}

		apiKey := parts[1]

		apiDetails, err := store.GetAPIKey(ctx, []byte(apiKey))
		if err != nil {
			log.Printf("Error: %s", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "please provide valid API Key"})
			ctx.Abort()
			return
		}

		err = token.VerifyAPIKey(apiKey, apiDetails.Signature, publicKey)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid api key",
			})
			ctx.Abort()
			return
		}

		ctx.Set(constants.PayloadKey, token.Payload{
			APIKey: apiKey,
			UserID: int(apiDetails.UserID),
		})

		ctx.Next()
	}
}
