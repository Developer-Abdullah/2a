package repository
import "platform/internal/store/postgres"
type cloneRepo struct { db *postgres.DB }
func NewCloneRepository(db *postgres.DB) *cloneRepo { return &cloneRepo{db: db} }
