package tenant
import "strings"
func SafeSchemaName(slug string) string {
return `"tenant_` + strings.ReplaceAll(strings.ReplaceAll(slug, "-", "_"), `"`, `""`) + `"`
}
