package postgres
import ( "net/url"; "github.com/golang-migrate/migrate/v4" )
func ApplyTenantMigrations(dbDSN, schemaName, migrationsPath string) error {
parsedURL, _ := url.Parse(dbDSN)
q := parsedURL.Query(); q.Set("search_path", schemaName); parsedURL.RawQuery = q.Encode()
m, err := migrate.New("file://"+migrationsPath, parsedURL.String())
if err != nil { return err }
defer m.Close()
if err := m.Up(); err != nil && err != migrate.ErrNoChange { return err }
return nil
}
