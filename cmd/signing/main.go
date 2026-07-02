package main

import (
"context"
"encoding/hex"
"os"
"os/signal"
"syscall"

"github.com/hibiken/asynq"
"github.com/rs/zerolog"
"github.com/rs/zerolog/log"

"platform/internal/repository"
"platform/internal/signing"
"platform/internal/store/postgres"
"platform/pkg/storage"
)

func main() {
zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

dbDSN := os.Getenv("DB_DSN")
redisURL := os.Getenv("REDIS_URL")
aesKeyHex := os.Getenv("AES_ENCRYPTION_KEY")

aesKey, err := hex.DecodeString(aesKeyHex)
if err != nil || len(aesKey) != 32 { log.Fatal().Msg("Invalid AES_ENCRYPTION_KEY") }

db, err := postgres.NewPool(dbDSN)
if err != nil { log.Fatal().Err(err).Msg("Database connection failed") }
defer db.Close()

s3Client, err := storage.NewS3Client(context.Background(), os.Getenv("S3_REGION"), os.Getenv("S3_ENDPOINT"), os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), os.Getenv("S3_BUCKET"))
if err != nil { log.Fatal().Err(err).Msg("S3 client failed") }

signingRepo := repository.NewSigningRepository(db)
signer := signing.ZsignSigner{Binary: os.Getenv("ZSIGN_BINARY")}
processor := signing.NewSigningProcessor(s3Client, signingRepo, signer, aesKey, os.Getenv("SIGNING_WORK_DIR"))

asynqOpts, _ := asynq.ParseRedisURI(redisURL)
asynqServer := asynq.NewServer(asynqOpts, asynq.Config{ Concurrency: 3, Queues: map[string]int{"critical": 6, "default": 3, "low": 1} })

mux := asynq.NewServeMux()
mux.Handle(signing.TaskIPASign, processor)

go func() {
if err := asynqServer.Run(mux); err != nil { log.Fatal().Err(err).Msg("Asynq server error") }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
asynqServer.Stop()
}
