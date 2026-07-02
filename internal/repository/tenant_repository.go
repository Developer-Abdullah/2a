package repository
import ( "context"; "database/sql"; "time"; "github.com/google/uuid"; "platform/internal/domain"; "platform/internal/store/postgres" )
type tenantRepo struct { db *postgres.DB }
func NewTenantRepository(db *postgres.DB) *tenantRepo { return &tenantRepo{db: db} }
func (r *tenantRepo) GetBySlugOrID(ctx context.Context, identifier string) (*domain.Tenant, error) {
query := `SELECT t.id, t.slug, t.plan_id, t.isolation_mode, t.schema_name, t.s3_prefix, t.status, t.created_at, p.name AS plan_name, p.max_apps, c.app_name FROM public.tenants t LEFT JOIN public.tenant_plans p ON t.plan_id = p.id LEFT JOIN public.tenant_config c ON t.id = c.tenant_id WHERE `
var dest struct { ID uuid.UUID `db:"id"`; Slug string `db:"slug"`; PlanID uuid.UUID `db:"plan_id"`; IsolationMode domain.IsolationMode `db:"isolation_mode"`; SchemaName string `db:"schema_name"`; S3Prefix string `db:"s3_prefix"`; Status domain.TenantStatus `db:"status"`; CreatedAt time.Time `db:"created_at"`; PlanName sql.NullString `db:"plan_name"`; MaxApps sql.NullInt64 `db:"max_apps"`; AppName sql.NullString `db:"app_name"` }
var err error
if parsedUUID, parseErr := uuid.Parse(identifier); parseErr == nil {
err = r.db.GetContext(ctx, &dest, query + `t.id = $1`, parsedUUID)
} else {
err = r.db.GetContext(ctx, &dest, query + `t.slug = $1`, identifier)
}
if err != nil { return nil, err }
t := &domain.Tenant{ ID: dest.ID, Slug: dest.Slug, PlanID: dest.PlanID, IsolationMode: dest.IsolationMode, SchemaName: dest.SchemaName, S3Prefix: dest.S3Prefix, Status: dest.Status, CreatedAt: dest.CreatedAt }
if dest.PlanName.Valid { t.Plan = &domain.TenantPlan{ ID: dest.PlanID, Name: dest.PlanName.String, MaxApps: int(dest.MaxApps.Int64) } }
if dest.AppName.Valid { t.Config = &domain.TenantConfig{ TenantID: dest.ID, AppName: dest.AppName.String } }
return t, nil
}
