package postgres
import ( "context"; "fmt"; "strings"; "time"; _ "github.com/jackc/pgx/v5/stdlib"; "github.com/jmoiron/sqlx" )
type DB struct { *sqlx.DB }
func NewPool(dsn string) (*DB, error) {
db, err := sqlx.Connect("pgx", dsn)
if err != nil { return nil, err }
db.SetMaxOpenConns(50); db.SetMaxIdleConns(10); db.SetConnMaxIdleTime(5 * time.Minute); db.SetConnMaxLifetime(30 * time.Minute)
if err := db.Ping(); err != nil { return nil, err }
return &DB{DB: db}, nil
}
func QuoteIdentifier(name string) string { return `"` + strings.ReplaceAll(name, `"`, `""`) + `"` }
func (db *DB) SetDedicatedSchema(ctx context.Context, tx *sqlx.Tx, schemaName string) error {
_, err := tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL search_path TO %s, public", QuoteIdentifier(schemaName)))
return err
}
