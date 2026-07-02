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
	"github.com/rs/zerolog/log"

	"platform/internal/admin"
	"platform/internal/auth"
	"platform/internal/store/postgres"
	"platform/pkg/logger"
	"platform/pkg/storage"
)

func main() {
	logger.InitLogger(true)
	log.Info().Msg("Starting Admin API Server...")

	db, err := postgres.NewPool(getEnvOrFatal("DB_DSN"))
	if err != nil {
		log.Fatal().Err(err).Msg("Database connection failed")
	}
	defer db.Close()

	privKey, err := auth.ParsePrivateKey(getEnvOrFatal("JWT_PRIVATE_KEY_PEM"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse ES256 private key")
	}
	pubKey, err := auth.ParsePublicKey(getEnvOrFatal("JWT_PUBLIC_KEY_PEM"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse ES256 public key")
	}

	s3Client, err := storage.NewS3Client(
		context.Background(),
		getEnv("S3_REGION", "us-east-1"),
		getEnvOrFatal("S3_ENDPOINT"),
		getEnvOrFatal("S3_ACCESS_KEY"),
		getEnvOrFatal("S3_SECRET_KEY"),
		getEnvOrFatal("S3_BUCKET"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("S3 client failed")
	}

	asynqOpts, err := asynq.ParseRedisURI(getEnvOrFatal("REDIS_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse Redis URL")
	}
	asynqClient := asynq.NewClient(asynqOpts)
	defer asynqClient.Close()

	router := admin.SetupAdminRouter(db, s3Client, asynqClient, privKey, pubKey)
	srv := &http.Server{Addr: ":8081", Handler: router, ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 120 * time.Second}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Admin API listening for requests")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Admin HTTP server crashed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Admin API forced to shutdown")
	}
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
