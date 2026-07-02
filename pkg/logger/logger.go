package logger
import ( "os"; "time"; "github.com/rs/zerolog"; "github.com/rs/zerolog/log" )

func InitLogger(debug bool) {
zerolog.TimeFieldFormat = time.RFC3339Nano
zerolog.SetGlobalLevel(zerolog.InfoLevel)
if debug { zerolog.SetGlobalLevel(zerolog.DebugLevel) }
log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05.000"}).With().Caller().Logger()
}
