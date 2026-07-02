package domain
import "github.com/google/uuid"
type UpdateType string
const ( UpdateTypeOptional UpdateType = "optional"; UpdateTypeForced UpdateType = "forced"; UpdateTypeSilent UpdateType = "silent" )
type Update struct { ID uuid.UUID `db:"id" json:"id"`; ApplicationID uuid.UUID `db:"application_id" json:"application_id"`; FromVersionID uuid.UUID `db:"from_version_id" json:"from_version_id"`; ToVersionID uuid.UUID `db:"to_version_id" json:"to_version_id"`; UpdateType UpdateType `db:"update_type" json:"update_type"` }
