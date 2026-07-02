package domain
import "github.com/google/uuid"
type ActivationCode struct { ID uuid.UUID `db:"id" json:"id"`; Code string `db:"code" json:"code"`; MaxDevices int `db:"max_devices" json:"max_devices"`; CurrentDeviceCount int `db:"current_device_count" json:"current_device_count"` }
