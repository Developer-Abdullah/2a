package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var validSchemaName = regexp.MustCompile(`^[a-z0-9_]+$`)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	target := flag.String("target", "", "Migration target: 'public' or 'tenant'")
	tenantID := flag.String("tenant-id", "", "Tenant slug")
	action := flag.String("action", "", "Action: 'up' or 'down'")
	dbDSN := flag.String("db", "", "PostgreSQL DSN")
	flag.Parse()

	if *dbDSN == "" {
		log.Fatal().Msg("-db is required")
	}
	if *action != "up" && *action != "down" {
		log.Fatal().Msg("-action must be 'up' or 'down'")
	}

	sourceURL := ""
	targetDSN := *dbDSN

	switch *target {
	case "public":
		sourceURL = "file://migrations/public"
	case "tenant":
		if *tenantID == "" {
			log.Fatal().Msg("-tenant-id is required for tenant migrations")
		}
		schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(*tenantID, "-", "_"))
		if !validSchemaName.MatchString(schemaName) {
			log.Fatal().Str("schema", schemaName).Msg("invalid tenant schema name")
		}
		if err := ensureSchema(*dbDSN, schemaName); err != nil {
			log.Fatal().Err(err).Str("schema", schemaName).Msg("failed to ensure tenant schema")
		}

		sourceURL = "file://migrations/tenant"
		parsedURL, err := url.Parse(targetDSN)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse database DSN")
		}
		q := parsedURL.Query()
		q.Set("search_path", schemaName)
		parsedURL.RawQuery = q.Encode()
		targetDSN = parsedURL.String()
	default:
		log.Fatal().Msg("-target must be 'public' or 'tenant'")
	}

	m, err := migrate.New(sourceURL, targetDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init migration")
	}
	defer m.Close()

	if *action == "up" {
		err = m.Up()
	} else {
		err = m.Down()
	}

	if err != nil {
		if err == migrate.ErrNoChange {
			log.Info().Msg("No new migrations to apply")
			return
		}
		log.Fatal().Err(err).Msg("Migration failed")
	}
	log.Info().Msg("Migrations applied successfully!")
}

func ensureSchema(dbDSN, schemaName string) error {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE SCHEMA IF NOT EXISTS ` + quoteIdentifier(schemaName))
	return err
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
