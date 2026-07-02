package domain

// NotificationDTO is the user-facing shape returned by GET /v1/notifications,
// joined against user_notification_reads so the client knows the read state.
type NotificationDTO struct {
	ID          string `db:"id" json:"id"`
	Title       string `db:"title" json:"title"`
	Body        string `db:"body" json:"body"`
	TargetAppID string `db:"target_app_id" json:"target_app_id"`
	SentAt      string `db:"sent_at" json:"sent_at"`
	ReadAt      string `db:"read_at" json:"read_at"`
}

// UpdateDTO mirrors the iOS Update model (camelCase JSON) for GET /v1/updates.
type UpdateDTO struct {
	ID             string `db:"id" json:"id"`
	AppID          string `db:"app_id" json:"appId"`
	CurrentVersion string `db:"current_version" json:"currentVersion"`
	LatestVersion  string `db:"latest_version" json:"latestVersion"`
	UpdateType     string `db:"update_type" json:"updateType"`
}

// VersionDTO is the admin-facing view of an application version (signed-apps list).
type VersionDTO struct {
	ID              string `db:"id" json:"id"`
	Version         string `db:"version" json:"version"`
	BuildNumber     string `db:"build_number" json:"build_number"`
	SigningStatus   string `db:"signing_status" json:"signing_status"`
	ReviewStatus    string `db:"review_status" json:"review_status"`
	RejectionReason string `db:"rejection_reason" json:"rejection_reason,omitempty"`
	SignedIPAS3Key  string `db:"signed_key" json:"signed_ipa_s3_key"`
	SizeBytes       int64  `db:"size_bytes" json:"size_bytes"`
	CreatedAt       string `db:"created_at" json:"created_at"`
}

// InstallTarget holds the data needed to render an OTA manifest and presign a download.
type InstallTarget struct {
	AppName   string `db:"app_name"`
	BundleID  string `db:"bundle_id"`
	Version   string `db:"version"`
	SignedKey string `db:"signed_key"`
	IconKey   string `db:"icon_key"`
}

// --- Admin dashboard read DTOs ---

type AdminNotificationDTO struct {
	ID     string `db:"id" json:"id"`
	Title  string `db:"title" json:"title"`
	Body   string `db:"body" json:"body"`
	Type   string `db:"type" json:"type"`
	SentAt string `db:"sent_at" json:"sent_at"`
}

type AdminCodeDTO struct {
	ID                 string `db:"id" json:"id"`
	Code               string `db:"code" json:"code"`
	Type               string `db:"type" json:"type"`
	DeviceType         string `db:"device_type" json:"device_type"`
	MaxDevices         int    `db:"max_devices" json:"max_devices"`
	CurrentDeviceCount int    `db:"current_device_count" json:"current_device_count"`
	MaxUses            int    `db:"max_uses" json:"max_uses"`
	CurrentUses        int    `db:"current_uses" json:"current_uses"`
	IsRevoked          bool   `db:"is_revoked" json:"is_revoked"`
}

type AdminUserDTO struct {
	ID          string `db:"id" json:"id"`
	DisplayName string `db:"display_name" json:"display_name"`
	DeviceCount int    `db:"device_count" json:"device_count"`
	CreatedAt   string `db:"created_at" json:"created_at"`
	LastSeenAt  string `db:"last_seen_at" json:"last_seen_at"`
}

type AdminDeviceDTO struct {
	ID               string `db:"id" json:"id"`
	UserID           string `db:"user_id" json:"user_id"`
	DeviceType       string `db:"device_type" json:"device_type"`
	Model            string `db:"model" json:"model"`
	EnrollmentMethod string `db:"enrollment_method" json:"enrollment_method"`
	IsRevoked        bool   `db:"is_revoked" json:"is_revoked"`
	LastSeenAt       string `db:"last_seen_at" json:"last_seen_at"`
}

type AdminStatsDTO struct {
	Apps          int     `db:"apps" json:"apps"`
	Versions      int     `db:"versions" json:"versions"`
	Users         int     `db:"users" json:"users"`
	Devices       int     `db:"devices" json:"devices"`
	Codes         int     `db:"codes" json:"codes"`
	Notifications int     `db:"notifications" json:"notifications"`
	AvgRating     float64 `db:"avg_rating" json:"avg_rating"`
	RatingCount   int     `db:"rating_count" json:"rating_count"`
}

type AdminSigningJobDTO struct {
	ID        string `db:"id" json:"id"`
	AppName   string `db:"app_name" json:"app_name"`
	Version   string `db:"version" json:"version"`
	Status    string `db:"status" json:"status"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

type AdminCertDTO struct {
	ID        string `db:"id" json:"id"`
	Label     string `db:"label" json:"label"`
	IsActive  bool   `db:"is_active" json:"is_active"`
	AddedAt   string `db:"added_at" json:"added_at"`
	ExpiresAt string `db:"expires_at" json:"expires_at"`
}

type AdminAuditDTO struct {
	ID           string `db:"id" json:"id"`
	Action       string `db:"action" json:"action"`
	ResourceType string `db:"resource_type" json:"resource_type"`
	ResourceID   string `db:"resource_id" json:"resource_id"`
	IPAddress    string `db:"ip_address" json:"ip_address"`
	CreatedAt    string `db:"created_at" json:"created_at"`
}

type AdminAccountDTO struct {
	ID        string `db:"id" json:"id"`
	Email     string `db:"email" json:"email"`
	Role      string `db:"role" json:"role"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

type AdminUserDetail struct {
	User    AdminUserDTO     `json:"user"`
	Devices []AdminDeviceDTO `json:"devices"`
}

// TenantSummaryDTO is the platform-owner view of a merchant.
type TenantSummaryDTO struct {
	Slug       string `db:"slug" json:"slug"`
	Name       string `db:"name" json:"name"`
	SchemaName string `db:"schema_name" json:"schema_name"`
	Status     string `db:"status" json:"status"`
	CreatedAt  string `db:"created_at" json:"created_at"`
	Apps       int    `json:"apps"`
}
