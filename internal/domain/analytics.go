package domain

import (
	"time"

	"github.com/google/uuid"
)

type AnalyticsDaily struct {
	Date          time.Time `db:"date" json:"date"`
	ApplicationID uuid.UUID `db:"application_id" json:"application_id"`
	VersionID     uuid.UUID `db:"version_id" json:"version_id"`
	Installs      int       `db:"installs" json:"installs"`
	Updates       int       `db:"updates" json:"updates"`
	Opens         int       `db:"opens" json:"opens"`
	Errors        int       `db:"errors" json:"errors"`
	UniqueUsers   int       `db:"unique_users" json:"unique_users"`
	UniqueDevices int       `db:"unique_devices" json:"unique_devices"`
}