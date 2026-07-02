package main

import (
	"context"
	"encoding/json"
	"net/smtp"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"platform/internal/billing"
	"platform/internal/notify"
	"platform/internal/repository"
	"platform/internal/store/postgres"
)

// renewalReminderDays is how far ahead of expiry a tenant is reminded to renew.
const renewalReminderDays = 7

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	log.Info().Msg("Starting Notifier worker...")

	asynqOpts, err := asynq.ParseRedisURI(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse Redis URI")
	}
	server := asynq.NewServer(asynqOpts, asynq.Config{Concurrency: 10, Queues: map[string]int{"notifications": 10}})

	mux := asynq.NewServeMux()
	mux.HandleFunc(notify.TypePush, handlePush)
	mux.HandleFunc(notify.TypeEmail, handleEmail)

	go func() {
		if err := server.Run(mux); err != nil {
			log.Fatal().Err(err).Msg("Notifier worker failed")
		}
	}()

	// Periodic billing maintenance: flag expiring subscriptions and advance lapsed statuses.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runBillingScheduler(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	server.Stop()
}

// runBillingScheduler runs the renewal-reminder + status-sweep pass hourly. It is a no-op-friendly
// loop: if the DB is unreachable it logs and retries on the next tick rather than crashing the worker.
func runBillingScheduler(ctx context.Context) {
	db, err := postgres.NewPool(os.Getenv("DB_DSN"))
	if err != nil {
		log.Error().Err(err).Msg("billing scheduler: DB connection failed; scheduler disabled")
		return
	}
	registry, graceDays := billing.FromEnv()
	svc := billing.NewService(repository.NewBillingRepository(db), registry, graceDays)

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	run := func() {
		if n, err := svc.SendRenewalReminders(ctx, renewalReminderDays); err != nil {
			log.Error().Err(err).Msg("renewal reminders failed")
		} else if n > 0 {
			log.Info().Int("count", n).Msg("queued renewal reminders")
		}
		if err := svc.SweepStatuses(ctx); err != nil {
			log.Error().Err(err).Msg("subscription status sweep failed")
		}
	}

	run() // once on startup
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

// handlePush dispatches a queued notification. Real APNs delivery (pkg/apns) needs per-device
// tokens and an APNs auth key; in this environment we log the dispatch so the pipeline is
// observable end-to-end rather than silently dropping the task.
func handlePush(ctx context.Context, t *asynq.Task) error {
	var p notify.PushPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Info().
		Str("notification_id", p.NotificationID).
		Str("tenant_id", p.TenantID).
		Str("title", p.Title).
		Msg("Dispatching push notification")
	return nil
}

// handleEmail delivers a transactional email (e.g. activation codes). It sends via SMTP when
// SMTP_HOST is configured; otherwise it logs the email so the pipeline is observable in dev without
// a mail server. A send failure returns an error so asynq retries the task.
func handleEmail(ctx context.Context, t *asynq.Task) error {
	var p notify.EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	host := os.Getenv("SMTP_HOST")
	if host == "" {
		// No mail server configured: log the delivery so it is visible end-to-end.
		log.Info().Str("to", p.To).Str("subject", p.Subject).Msg("Email delivery (SMTP not configured; logged only)")
		return nil
	}

	port := getEnvDefault("SMTP_PORT", "587")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	// Many providers (Gmail especially) require the envelope From to be the authenticated user, so
	// default SMTP_FROM to SMTP_USER when it is not set explicitly.
	from := getEnvDefault("SMTP_FROM", firstNonEmpty(user, "no-reply@store.local"))

	msg := buildMIME(from, p.To, p.Subject, p.Body)
	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	if err := smtp.SendMail(addr, auth, from, []string{p.To}, msg); err != nil {
		log.Error().Err(err).Str("to", p.To).Msg("SMTP send failed")
		return err
	}
	log.Info().Str("to", p.To).Str("subject", p.Subject).Msg("Email delivered")
	return nil
}

// buildMIME renders a minimal UTF-8 text/plain email with the standard headers.
func buildMIME(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
