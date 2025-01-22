package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	database "github.com/DEVunderdog/transcript-generator-backend/database/sqlc"
	"github.com/DEVunderdog/transcript-generator-backend/logger"
	"github.com/DEVunderdog/transcript-generator-backend/server"
	"github.com/DEVunderdog/transcript-generator-backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logConfig := logger.LoggerConfig{
		LogLevel: zerolog.InfoLevel,
		FileConfig: &logger.FileConfig{
			Path:       "../logs/backend/backend.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
	}

	baseLogger := logger.NewLogger(logConfig)

	config, err := utils.LoadConfig("../.env/backend.env")
	if err != nil {
		baseLogger.Logger.Fatal().Err(err).Msg("error loading configuration")
	}

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		baseLogger.Logger.Fatal().Err(err).Msg("error creating database connection pool")
	}

	store := database.NewStore(connPool)

	goth.UseProviders(
		google.New(config.ClientID, config.ClientSecret, config.RedirectURL),
	)

	server, err := server.NewServer(ctx, store, config, baseLogger)
	if err != nil {
		baseLogger.Logger.Fatal().Err(err).Msg("error creating server")
	}

	srv := server.Start()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			baseLogger.Logger.Fatal().Err(err).Msg("error in server while listening")
		}
	}()

	<-ctx.Done()

	stop()

	// if err := srv.Shutdown(ctx); err != nil {
	// 	baseLogger.Logger.Fatal().Msg("server shutting down")
	// }

	if err := server.ServerShutdown(ctx, srv); err != nil {
		baseLogger.Logger.Fatal().Err(err).Msg("error while server is shutting down")
	}

	baseLogger.Logger.Info().Msg("Bye :)")
}
