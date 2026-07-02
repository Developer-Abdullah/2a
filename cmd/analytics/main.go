package main

import (
"os"
"os/signal"
"syscall"
"github.com/hibiken/asynq"
"github.com/rs/zerolog"
"github.com/rs/zerolog/log"
)

func main() {
zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

asynqOpts, _ := asynq.ParseRedisURI(os.Getenv("REDIS_URL"))
server := asynq.NewServer(asynqOpts, asynq.Config{ Concurrency: 2, Queues: map[string]int{"low": 10} })
mux := asynq.NewServeMux()

go func() {
if err := server.Run(mux); err != nil { log.Fatal().Err(err).Msg("Analytics worker failed") }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
server.Stop()
}
