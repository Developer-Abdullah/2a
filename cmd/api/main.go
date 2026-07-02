package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"platform/internal/api"
	"platform/internal/auth"
	"platform/internal/store/postgres"
	"platform/pkg/storage"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	log.Info().Msg("Starting Core API Server...")

	dbDSN := getEnvOrFatal("DB_DSN")
	redisURL := getEnvOrFatal("REDIS_URL")
	privKeyPEM := getEnvOrFatal("JWT_PRIVATE_KEY_PEM")
	pubKeyPEM := getEnvOrFatal("JWT_PUBLIC_KEY_PEM")

	privKey, err := auth.ParsePrivateKey(privKeyPEM)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse ES256 Private Key")
	}

	pubKey, err := auth.ParsePublicKey(pubKeyPEM)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse ES256 Public Key")
	}

	db, err := postgres.NewPool(dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize PostgreSQL pool")
	}
	defer db.Close()

	s3Client, err := storage.NewS3Client(
		context.Background(),
		getEnv("S3_REGION", "us-east-1"),
		getEnv("S3_ENDPOINT", ""),
		getEnvOrFatal("S3_ACCESS_KEY"),
		getEnvOrFatal("S3_SECRET_KEY"),
		getEnvOrFatal("S3_BUCKET"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize S3 client")
	}

	asynqOpts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse Redis URI for Asynq")
	}
	asynqClient := asynq.NewClient(asynqOpts)
	defer asynqClient.Close()

	router := api.SetupRouter(db, s3Client, asynqClient, privKey, pubKey)
	srv := &http.Server{Addr: ":8080", Handler: router, ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 120 * time.Second}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Core API listening for requests")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP server encountered a fatal error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown abnormally")
	}
	log.Info().Msg("Core API exited cleanly.")
}

func getEnvOrFatal(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal().Str("env_var", key).Msg("Missing required environment variable")
	}
	return val
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
