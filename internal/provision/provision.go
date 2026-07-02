package provision

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"platform/internal/store/postgres"
	"platform/migrations"
)

// ProvisionTenantSchema creates the tenant schema (if absent) and applies every tenant migration
// to it, on a SINGLE pinned connection so the search_path holds for the whole run. This is the
// runtime equivalent of `migrate -target=tenant`, used by the platform owner to onboard a merchant.
func ProvisionTenantSchema(ctx context.Context, db *postgres.DB, schemaName string) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	quoted := postgres.QuoteIdentifier(schemaName)
	if _, err := conn.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+quoted); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", quoted)); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrations.TenantFiles, "tenant")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		content, err := fs.ReadFile(migrations.TenantFiles, "tenant/"+name)
		if err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}
