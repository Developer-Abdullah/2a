package domain
import ( "time"; "github.com/google/uuid" )
type TenantAdminRole string
const ( RoleSuperAdmin TenantAdminRole = "super_admin"; RoleAdmin TenantAdminRole = "admin"; RoleSupport TenantAdminRole = "support" )
type AdminUser struct {
ID uuid.UUID `db:"id" json:"id"`
Email string `db:"email" json:"email"`
PasswordHash string `db:"password_hash" json:"-"`
Role TenantAdminRole `db:"role" json:"role"`
TenantSlug string `db:"tenant_slug" json:"tenant_slug"`
IsActive bool `db:"is_active" json:"is_active"`
LastLoginAt *time.Time `db:"last_login_at" json:"last_login_at"`
CreatedAt time.Time `db:"created_at" json:"created_at"`
}
