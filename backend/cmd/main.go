package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	_ "github.com/DEVunderdog/transcript-generator-backend/docs"
	"github.com/DEVunderdog/transcript-generator-backend/internal/api"
	database "github.com/DEVunderdog/transcript-generator-backend/internal/database/sqlc"
	"github.com/DEVunderdog/transcript-generator-backend/internal/logger"
	"github.com/DEVunderdog/transcript-generator-backend/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// @title Transcript Generator API
// @version 1.0
// @description API for generating transcript from audio files using OpenAI Whisper Model, please note that you will receive the transcript.pdf file on your registered email address.
// @host transcript-generator-backend-29185933434.asia-south1.run.app
// @BasePath /server
// @schemes https
// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logConfig := logger.LoggerConfig{
		LogLevel: zerolog.InfoLevel,
	}

	baseLogger := logger.NewLogger(logConfig)

	config, err := utils.LoadProdConfig()
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("error loading configuration")
	}

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("error creating database connection pool")
	}

	store := database.NewStore(connPool)

	server, err := api.NewServer(ctx, store, config, baseLogger)
	if err != nil {
		baseLogger.Fatal().Err(err).Msg("error creating server")
	}

	srv := server.Start()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			baseLogger.Fatal().Err(err).Msg("error in server while listening")
		}
	}()

	<-ctx.Done()

	stop()

	if err := server.ServerShutdown(ctx, srv); err != nil {
		baseLogger.Fatal().Err(err).Msg("error while server is shutting down")
	}

	baseLogger.Info().Msg("Bye :)")
}
