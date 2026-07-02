package migrations

import "embed"

// TenantFiles embeds the tenant migration SQL so the platform (owner) service can provision a new
// merchant's schema at runtime — the same DDL `migrate -target=tenant` applies from disk.
//
//go:embed tenant/*.sql
var TenantFiles embed.FS
