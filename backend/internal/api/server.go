package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	database "github.com/DEVunderdog/transcript-generator-backend/internal/database/sqlc"
	"github.com/DEVunderdog/transcript-generator-backend/internal/gcp/cloud_pubsub"
	"github.com/DEVunderdog/transcript-generator-backend/internal/gcp/storage"
	"github.com/DEVunderdog/transcript-generator-backend/internal/logger"
	"github.com/DEVunderdog/transcript-generator-backend/internal/middleware"

	"github.com/DEVunderdog/transcript-generator-backend/internal/token"
	"github.com/DEVunderdog/transcript-generator-backend/internal/utils"

	_ "github.com/DEVunderdog/transcript-generator-backend/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router        *gin.Engine
	config        *utils.Config
	store         database.Store
	tokenMaker    token.TokenMaker
	storageClient *storage.StorageClient
	pubsubClient  *cloud_pubsub.CloudPubSubClient
	baseLogger    *logger.Logger
	httpLogger    *middleware.HTTPLogger
}

// @Description Response data structure
type responseData struct {
	Status int `json:"status"`
	Data   any `json:"data,omitempty"`
}

// @Description Standard response structure
type standardResponse struct {
	Message  string       `json:"message"`
	Response responseData `json:"response"`
}

func (server *Server) enhanceHTTPResponse(ctx *gin.Context, httpStatus int, message string, data any) {
	response := standardResponse{
		Message: message,
		Response: responseData{
			Status: httpStatus,
		},
	}

	if data != nil {
		response.Response.Data = data
	}

	ctx.JSON(httpStatus, response)
}

func (server *Server) Start() *http.Server {
	srv := &http.Server{
		Addr:    ":" + server.config.Port,
		Handler: server.router,
	}

	return srv
}


func NewServer(ctx context.Context, store database.Store, config *utils.Config, baseLogger *logger.Logger) (*Server, error) {
	httpLogger := middleware.NewHTTPLogger(baseLogger)

	err := token.InitializeJWTKeys(config.Passphrase, store, ctx, config.KeysPurpose)
	if err != nil {
		return nil, fmt.Errorf("error initializing encryption keys: %w", err)
	}

	jwtKeyResponse, err := token.GetKeyBasedOnPurpose(ctx, store, config.KeysPurpose)
	if err != nil {
		return nil, fmt.Errorf("error getting encryption keys: %w", err)
	}

	privateKey, err := token.GetPrivateKey(jwtKeyResponse.PrivateKey, []byte(config.Passphrase))
	if err != nil {
		return nil, fmt.Errorf("error getting private key: %w", err)
	}

	publicKey, err := token.GetPublicKey([]byte(jwtKeyResponse.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("error getting public key: %w", err)
	}

	tokenMaker := token.NewTokenMaker(publicKey, privateKey)

	storageClient, err := storage.NewStorageClient(ctx, config.ServiceAccountKeyPath, config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("error creating storage client in gcp: %w", err)
	}

	pubSubClient, err := cloud_pubsub.NewCloudPubSubClient(ctx, config.ServiceAccountKeyPath, config.TopicID, config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("error while creating pub sub client in gcp: %w", err)
	}

	server := &Server{
		config:        config,
		store:         store,
		tokenMaker:    *tokenMaker,
		storageClient: storageClient,
		pubsubClient:  pubSubClient,
		baseLogger:    baseLogger,
		httpLogger:    httpLogger,
	}

	if err := server.setupRouter(); err != nil {
		return nil, fmt.Errorf("error setting up router: %w", err)
	}

	return server, nil
}

func (server *Server) ServerShutdown(ctx context.Context, srv *http.Server) error {

	if err := server.storageClient.Close(); err != nil {
		return fmt.Errorf("error closing gcp storage client: %w", err)
	}

	if err := server.pubsubClient.Close(); err != nil {
		return fmt.Errorf("error closing gcp pub sub client: %w", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down http server: %w", err)
	}

	return nil
}

func (server *Server) setupRouter() error {
	router := gin.Default()

	router.SetTrustedProxies(nil)

	router.ForwardedByClientIP = true

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true, // needs to change in prod
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:   []string{"Content-Length"},
		MaxAge:          24 * time.Hour,
	}))

	router.Use(middleware.SecurityHeaderMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(server.httpLogger.LoggingMiddleware())

	router.GET("/server/health", server.serverHealthCheck)
	router.POST("/server/api/register", server.generateAPIKey)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authRoutes := router.Group("/server/auth")
	{
		authRoutes.Use(middleware.Authenticate(*server.config, server.store))
		authRoutes.DELETE("/api/delete", server.deleteAPIKey)
	}

	fileRoutes := authRoutes.Group("/files")
	{
		fileRoutes.POST("/upload", server.uploadFileToBucket)
		fileRoutes.POST("/update", server.updateFile)
		fileRoutes.GET("/list", server.listAllFiles)
		fileRoutes.DELETE("/delete/:filename", server.deleteFile)
		fileRoutes.GET("/sync", server.sync)
	}

	transcriptRoutes := authRoutes.Group("/transcript")
	{
		transcriptRoutes.GET("/request", server.requestTranscript)
	}

	server.router = router

	return nil
}
