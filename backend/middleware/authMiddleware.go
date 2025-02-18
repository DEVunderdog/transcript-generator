package middleware

import (
	"log"
	"net/http"

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

		apiDetails, err := store.GetAPIKey(ctx, []byte(authHeader))
		if err != nil {
			log.Printf("Error: %s", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "please provide valid API Key"})
			ctx.Abort()
			return
		}

		err = token.VerifyAPIKey(authHeader, apiDetails.Signature, publicKey)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid api key",
			})
			ctx.Abort()
			return
		}

		ctx.Set(constants.PayloadKey, token.Payload{
			APIKey: authHeader,
			UserID: int(apiDetails.UserID),
		})

		ctx.Next()
	}
}
