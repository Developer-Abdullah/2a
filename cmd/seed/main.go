package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"platform/internal/store/postgres"
	"platform/pkg/logger"
)

func main() {
	logger.InitLogger(true)
	db, err := postgres.NewPool(getEnvDSN())
	if err != nil {
		log.Fatal().Err(err).Msg("DB connection failed")
	}
	defer db.Close()

	ctx := context.Background()

	tenantID := seedPublic(ctx, db)

	// Tenant tables are created by the tenant migration, which runs after the first seed pass.
	// We seed tenant data only once those tables exist, so this command is safe to run twice
	// (before and after `migrate -target=tenant`).
	if tenantTablesExist(ctx, db) {
		seedTenant(ctx, db, tenantID)
		log.Info().Msg("Tenant demo data seeded (apps, versions, user, codes, ratings, notification)")
	} else {
		log.Info().Msg("Tenant tables not present yet; skipping tenant data (re-run seed after tenant migration)")
	}

	log.Info().Str("tenant_id", tenantID).Msg("Seeding completed successfully")
}

func getEnvDSN() string {
	// Mirrors the original behaviour: DB_DSN drives the connection.
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal().Str("env_var", "DB_DSN").Msg("Missing required environment variable")
	}
	return dsn
}

// seedPublic provisions the platform-level rows: plan, tenant, config, schema, admin, certificate.
func seedPublic(ctx context.Context, db *postgres.DB) string {
	var planID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO public.tenant_plans
			(name, max_apps, max_users, max_storage_gb, max_devices_per_code, isolation_mode,
			 price_amount, currency, billing_interval)
		VALUES
			('Enterprise Unlimited', 9999, 999999, 5000.0, 5, 'dedicated_schema',
			 5.000, 'KWD', 'yearly')
		ON CONFLICT (name) DO UPDATE SET
			price_amount = EXCLUDED.price_amount,
			currency = EXCLUDED.currency,
			billing_interval = EXCLUDED.billing_interval
		RETURNING id
	`).Scan(&planID); err != nil {
		log.Fatal().Err(err).Msg("failed to seed tenant plan")
	}

	var tenantID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO public.tenants
			(slug, plan_id, isolation_mode, schema_name, s3_prefix)
		VALUES
			('store', $1, 'dedicated_schema', 'tenant_store', 'tenant_store/')
		ON CONFLICT (slug) DO UPDATE SET
			plan_id = EXCLUDED.plan_id,
			isolation_mode = EXCLUDED.isolation_mode,
			schema_name = EXCLUDED.schema_name,
			s3_prefix = EXCLUDED.s3_prefix
		RETURNING id
	`, planID).Scan(&tenantID); err != nil {
		log.Fatal().Err(err).Msg("failed to seed tenant")
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO public.tenant_config (tenant_id, app_name, support_email)
		VALUES ($1, 'Double A', 'support@platform.local')
		ON CONFLICT (tenant_id) DO UPDATE SET
			app_name = EXCLUDED.app_name,
			support_email = EXCLUDED.support_email,
			updated_at = NOW()
	`, tenantID); err != nil {
		log.Fatal().Err(err).Msg("failed to seed tenant config")
	}

	if _, err := db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS tenant_store`); err != nil {
		log.Fatal().Err(err).Msg("failed to create tenant schema")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to hash default admin password")
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO public.platform_admins (email, password_hash, role)
		VALUES ('admin@platform.com', $1, 'super_admin')
		ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role
	`, string(hash)); err != nil {
		log.Fatal().Err(err).Msg("failed to seed platform admin")
	}

	// An active signing certificate is required before the admin "Upload & Sign" flow can
	// enqueue a job (signing_jobs.certificate_id is NOT NULL).
	if _, err := db.ExecContext(ctx, `
		INSERT INTO public.tenant_certificates
			(tenant_id, label, s3_key_p12_enc, s3_key_mobileprovision, encrypted_password, expires_at)
		SELECT $1, 'Demo Enterprise Cert', 'demo/cert.p12.enc', 'demo/profile.mobileprovision', $2, NOW() + INTERVAL '365 days'
		WHERE NOT EXISTS (SELECT 1 FROM public.tenant_certificates WHERE tenant_id = $1)
	`, tenantID, []byte("demo-encrypted-password")); err != nil {
		log.Fatal().Err(err).Msg("failed to seed signing certificate")
	}

	return tenantID
}

func tenantTablesExist(ctx context.Context, db *postgres.DB) bool {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('tenant_store.applications') IS NOT NULL`).Scan(&exists); err != nil {
		return false
	}
	return exists
}

// seedTenant fills the tenant_store schema with demonstrable data. It runs inside a single
// transaction with search_path set so the enum casts (::app_category, etc.) resolve, and every
// insert is idempotent so repeated runs are harmless.
func seedTenant(ctx context.Context, db *postgres.DB, tenantID string) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to begin tenant seed tx")
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO tenant_store, public`); err != nil {
		log.Fatal().Err(err).Msg("failed to set search_path")
	}

	type appSeed struct {
		Bundle, Name, Category, Desc, Features string
	}
	apps := []appSeed{
		{"com.demo.cloudnotes", "Cloud Notes", "productivity", "Sync your notes across every device.", `{"Realtime sync","End-to-end encryption","Markdown support"}`},
		{"com.demo.pulsemusic", "Pulse Music", "entertainment", "Stream millions of tracks offline.", `{"Offline mode","Hi-fi audio","Smart playlists"}`},
		{"com.demo.shieldvpn", "Shield VPN", "other", "Private, fast and secure browsing.", `{"No-logs policy","50+ countries","Kill switch"}`},
	}

	appIDs := make(map[string]string, len(apps))
	for _, a := range apps {
		var id string
		if err := tx.QueryRowxContext(ctx, `
			INSERT INTO applications (bundle_identifier, name, category, description, features, is_published)
			VALUES ($1, $2, $3::app_category, $4, $5::text[], true)
			ON CONFLICT (bundle_identifier) DO UPDATE SET name = EXCLUDED.name, is_published = true
			RETURNING id::text
		`, a.Bundle, a.Name, a.Category, a.Desc, a.Features).Scan(&id); err != nil {
			log.Fatal().Err(err).Str("app", a.Bundle).Msg("failed to seed application")
		}
		appIDs[a.Bundle] = id

		iconKey := fmt.Sprintf("%s/icons/%s.png", tenantID, id)
		if _, err := tx.ExecContext(ctx, `UPDATE applications SET icon_s3_key = $2 WHERE id = $1::uuid`, id, iconKey); err != nil {
			log.Fatal().Err(err).Msg("failed to set app icon key")
		}
	}

	ensureVersion := func(appID, version, build string, size int64) string {
		var id string
		err := tx.QueryRowxContext(ctx, `SELECT id::text FROM application_versions WHERE application_id = $1::uuid AND version = $2`, appID, version).Scan(&id)
		if err == nil {
			return id
		}
		if err != sql.ErrNoRows {
			log.Fatal().Err(err).Msg("failed to look up version")
		}
		signedKey := fmt.Sprintf("%s/signed-ipa/%s/%s.ipa", tenantID, appID, version)
		if err := tx.QueryRowxContext(ctx, `
			INSERT INTO application_versions
				(application_id, version, build_number, size_bytes, raw_ipa_s3_key, signed_ipa_s3_key, signing_status, published_at)
			VALUES ($1::uuid, $2, $3, $4, $5, $5, 'signed', NOW())
			RETURNING id::text
		`, appID, version, build, size, signedKey).Scan(&id); err != nil {
			log.Fatal().Err(err).Msg("failed to seed application version")
		}
		return id
	}

	// Cloud Notes gets two signed versions plus an update transition so the update feed/badge works.
	cloudID := appIDs["com.demo.cloudnotes"]
	v100 := ensureVersion(cloudID, "1.0.0", "100", 25_000_000)
	v110 := ensureVersion(cloudID, "1.1.0", "110", 26_500_000)
	ensureVersion(appIDs["com.demo.pulsemusic"], "1.0.0", "100", 48_000_000)
	ensureVersion(appIDs["com.demo.shieldvpn"], "1.0.0", "100", 18_000_000)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO updates (application_id, from_version_id, to_version_id, update_type)
		SELECT $1::uuid, $2::uuid, $3::uuid, 'optional'
		WHERE NOT EXISTS (
			SELECT 1 FROM updates WHERE application_id = $1::uuid AND from_version_id = $2::uuid AND to_version_id = $3::uuid
		)
	`, cloudID, v100, v110); err != nil {
		log.Fatal().Err(err).Msg("failed to seed update")
	}

	// Activation code + an activated user and bound device.
	var codeID string
	if err := tx.QueryRowxContext(ctx, `
		INSERT INTO activation_codes (code, type, device_type, max_devices, max_uses, current_device_count, current_uses)
		VALUES ('DEMO-CODE-1234', 'usage_count', 'both', 5, 100, 1, 1)
		ON CONFLICT (code) DO UPDATE SET max_uses = EXCLUDED.max_uses
		RETURNING id::text
	`).Scan(&codeID); err != nil {
		log.Fatal().Err(err).Msg("failed to seed activation code")
	}

	var userID string
	err = tx.QueryRowxContext(ctx, `SELECT id::text FROM users WHERE activation_code_id = $1::uuid LIMIT 1`, codeID).Scan(&userID)
	if err == sql.ErrNoRows {
		if err := tx.QueryRowxContext(ctx, `
			INSERT INTO users (activation_code_id, display_name, last_seen_at)
			VALUES ($1::uuid, 'Demo User', NOW())
			RETURNING id::text
		`, codeID).Scan(&userID); err != nil {
			log.Fatal().Err(err).Msg("failed to seed user")
		}
	} else if err != nil {
		log.Fatal().Err(err).Msg("failed to look up seed user")
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO devices (user_id, enrollment_method, fingerprint_hash, device_type, model_string, last_seen_at, last_validated_at)
		VALUES ($1::uuid, 'fingerprint', 'demo-fingerprint-hash', 'iphone', 'iPhone 15 Pro', NOW(), NOW())
		ON CONFLICT (fingerprint_hash) DO NOTHING
	`, userID); err != nil {
		log.Fatal().Err(err).Msg("failed to seed device")
	}

	// A few ratings (5, 4, 5) -> average 4.67 over 3 ratings. Distinct ip_hash avoids the daily unique index.
	ratings := []struct {
		Stars   int
		Comment string
		IPHash  string
	}{
		{5, "Best store ever!", "seed-ip-hash-1"},
		{4, "Works great, would love dark mode.", "seed-ip-hash-2"},
		{5, "Super fast installs.", "seed-ip-hash-3"},
	}
	for _, r := range ratings {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO platform_ratings (user_id, rating, comment, ip_hash)
			VALUES ($1::uuid, $2, $3, $4)
			ON CONFLICT (ip_hash, rating_date) DO NOTHING
		`, userID, r.Stars, r.Comment, r.IPHash); err != nil {
			log.Fatal().Err(err).Msg("failed to seed rating")
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO notifications (title, body, type)
		SELECT 'Welcome to the Store', 'Browse and install your favourite apps in one tap.', 'alert'
		WHERE NOT EXISTS (SELECT 1 FROM notifications WHERE title = 'Welcome to the Store')
	`); err != nil {
		log.Fatal().Err(err).Msg("failed to seed notification")
	}

	seedProducts(ctx, tx)

	if err := tx.Commit(); err != nil {
		log.Fatal().Err(err).Msg("failed to commit tenant seed")
	}
}

// seedProducts fills the storefront catalog with a couple of demonstrable subscription products, each
// priced in EGP and KWD. Idempotent on slug so repeated runs are harmless.
func seedProducts(ctx context.Context, tx *sqlx.Tx) {
	type price struct {
		Currency  string
		Amount    float64
		CompareAt float64 // 0 = no struck-through price
	}
	type productSeed struct {
		Slug             string
		Name             string
		Subtitle         string
		Description      string
		DeviceType       string
		SubscriptionDays int
		Features         string // JSON array
		Terms            string // JSON array
		VideoURL         string
		PurchaseCount    int
		RatingAvg        float64
		RatingCount      int
		SortOrder        int
		Prices           []price
	}

	products := []productSeed{
		{
			Slug:             "PdDWGXW",
			Name:             "اشتراك تطبيقات بلس — سنة كاملة (آيفون)",
			Subtitle:         "كود تفعيل فوري لأكثر من 9000 تطبيق",
			Description:      "اشتراك سنوي يمنحك الوصول إلى مكتبة تطبيقات بلس المعدّلة بدون جيلبريك، مع تفعيل فوري وضمان.",
			DeviceType:       "iphone",
			SubscriptionDays: 365,
			Features:         `["تفعيل فوري خلال دقائق","بدون جيلبريك","أكثر من 9000 تطبيق","دعم فني على مدار الساعة","ضمان استبدال"]`,
			Terms:            `["الكود غير قابل للاسترجاع بعد التفعيل","صلاحية الاشتراك سنة من تاريخ التفعيل","ضمان شهر على التفعيل"]`,
			VideoURL:         "",
			PurchaseCount:    10000,
			RatingAvg:        4.99,
			RatingCount:      72,
			SortOrder:        1,
			Prices: []price{
				{Currency: "EGP", Amount: 499.000, CompareAt: 1200.000},
				{Currency: "KWD", Amount: 3.294, CompareAt: 8.238},
			},
		},
		{
			Slug:             "IPADPLUS",
			Name:             "اشتراك تطبيقات بلس — سنة كاملة (آيباد)",
			Subtitle:         "كود تفعيل فوري للآيباد",
			Description:      "نفس مكتبة تطبيقات بلس محسّنة لأجهزة الآيباد، تفعيل فوري وضمان.",
			DeviceType:       "ipad",
			SubscriptionDays: 365,
			Features:         `["تفعيل فوري","بدون جيلبريك","محسّن للآيباد","دعم فني","ضمان استبدال"]`,
			Terms:            `["الكود غير قابل للاسترجاع بعد التفعيل","صلاحية الاشتراك سنة من تاريخ التفعيل"]`,
			VideoURL:         "",
			PurchaseCount:    3200,
			RatingAvg:        4.95,
			RatingCount:      38,
			SortOrder:        2,
			Prices: []price{
				{Currency: "EGP", Amount: 299.000, CompareAt: 700.000},
				{Currency: "KWD", Amount: 1.980, CompareAt: 4.950},
			},
		},
	}

	for _, p := range products {
		var productID string
		// rating_avg/rating_count are set only on initial insert (as starting display values). They are
		// NOT overwritten on re-seed, because a DB trigger keeps them in sync with real product_ratings —
		// re-clobbering them here would wipe genuine customer ratings.
		if err := tx.QueryRowxContext(ctx, `
			INSERT INTO products
				(slug, name, subtitle, description, device_type, subscription_days,
				 features, terms, video_url, is_published, purchase_count, rating_avg, rating_count, sort_order)
			VALUES ($1, $2, $3, $4, $5::allowed_device_type, $6, $7::jsonb, $8::jsonb, NULLIF($9,''), true, $10, $11, $12, $13)
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name, subtitle = EXCLUDED.subtitle, description = EXCLUDED.description,
				device_type = EXCLUDED.device_type, subscription_days = EXCLUDED.subscription_days,
				features = EXCLUDED.features, terms = EXCLUDED.terms, is_published = true,
				purchase_count = EXCLUDED.purchase_count, sort_order = EXCLUDED.sort_order, updated_at = NOW()
			RETURNING id::text
		`, p.Slug, p.Name, p.Subtitle, p.Description, p.DeviceType, p.SubscriptionDays,
			p.Features, p.Terms, p.VideoURL, p.PurchaseCount, p.RatingAvg, p.RatingCount, p.SortOrder).Scan(&productID); err != nil {
			log.Fatal().Err(err).Str("slug", p.Slug).Msg("failed to seed product")
		}

		for _, pr := range p.Prices {
			var compareAt interface{}
			if pr.CompareAt > 0 {
				compareAt = pr.CompareAt
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO product_prices (product_id, currency, amount, compare_at_amount)
				VALUES ($1::uuid, $2, $3, $4)
				ON CONFLICT (product_id, currency) DO UPDATE SET
					amount = EXCLUDED.amount, compare_at_amount = EXCLUDED.compare_at_amount
			`, productID, pr.Currency, pr.Amount, compareAt); err != nil {
				log.Fatal().Err(err).Str("slug", p.Slug).Str("currency", pr.Currency).Msg("failed to seed product price")
			}
		}
	}
}
