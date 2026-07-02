package repository

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"platform/internal/domain"
	"platform/internal/store/postgres"
)

type AdminRepository interface {
	AuthenticateAdmin(ctx context.Context, email, password string) (*domain.AdminUser, error)
}
type adminRepo struct{ db *postgres.DB }

func NewAdminRepository(db *postgres.DB) AdminRepository { return &adminRepo{db: db} }
func (r *adminRepo) AuthenticateAdmin(ctx context.Context, email, password string) (*domain.AdminUser, error) {
	query := `SELECT a.id, a.email, a.password_hash, a.role, COALESCE(t.slug, '') AS tenant_slug
		FROM public.platform_admins a
		LEFT JOIN public.tenants t ON t.id = a.tenant_id
		WHERE a.email = $1`
	var user domain.AdminUser
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}
