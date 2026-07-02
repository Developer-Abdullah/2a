package domain

import (
	"github.com/google/uuid"
	"time"
)

type IsolationMode string
type TenantStatus string
type EnrollmentMethod string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
)

type Tenant struct {
	ID            uuid.UUID     `db:"id" json:"id"`
	Slug          string        `db:"slug" json:"slug"`
	PlanID        uuid.UUID     `db:"plan_id" json:"plan_id"`
	IsolationMode IsolationMode `db:"isolation_mode" json:"isolation_mode"`
	SchemaName    string        `db:"schema_name" json:"schema_name"`
	S3Prefix      string        `db:"s3_prefix" json:"s3_prefix"`
	Status        TenantStatus  `db:"status" json:"status"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
	Plan          *TenantPlan   `db:"-" json:"plan,omitempty"`
	Config        *TenantConfig `db:"-" json:"config,omitempty"`
}
type TenantPlan struct {
	ID      uuid.UUID `db:"id" json:"id"`
	Name    string    `db:"name" json:"name"`
	MaxApps int       `db:"max_apps" json:"max_apps"`
}
type TenantConfig struct {
	TenantID uuid.UUID `db:"tenant_id" json:"-"`
	AppName  string    `db:"app_name" json:"app_name"`
}
