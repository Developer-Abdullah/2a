package repository
import "platform/internal/store/postgres"
type analyticsRepo struct { db *postgres.DB }
func NewAnalyticsRepository(db *postgres.DB) *analyticsRepo { return &analyticsRepo{db: db} }
