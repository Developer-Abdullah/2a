package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

var errProductUnavailable = errors.New("product is not available in this currency")

type shopRepo struct {
	db *postgres.DB
}

func NewShopRepository(db *postgres.DB) *shopRepo {
	return &shopRepo{db: db}
}

// ListPublishedProducts returns published products that have a price in the requested currency,
// each carrying that single currency's price.
func (r *shopRepo) ListPublishedProducts(ctx context.Context, tenantID uuid.UUID, currency string) ([]domain.Product, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.slug, p.name, COALESCE(p.subtitle,'') AS subtitle, COALESCE(p.description,'') AS description,
			p.device_type::text AS device_type, p.subscription_days, p.codes_per_unit,
			p.features, p.terms, COALESCE(p.video_url,'') AS video_url, COALESCE(p.image_s3_key,'') AS image_url,
			p.is_published, p.purchase_count, p.rating_avg, p.rating_count, p.sort_order, p.created_at,
			pr.currency, pr.amount, pr.compare_at_amount
		FROM %[1]s.products p
		JOIN %[1]s.product_prices pr ON pr.product_id = p.id AND pr.currency = $1
		WHERE p.is_published = true
		ORDER BY p.sort_order ASC, p.created_at DESC
	`, postgres.QuoteIdentifier(schema))

	rows, err := r.db.QueryxContext(ctx, query, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		p, price, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		p.Prices = []domain.ProductPrice{price}
		products = append(products, *p)
	}
	return products, rows.Err()
}

// GetProductBySlug returns a single published product with ALL its currency prices, so the
// storefront can switch currency on the product page without a refetch. The `currency` argument is
// accepted for symmetry but does not filter prices.
func (r *shopRepo) GetProductBySlug(ctx context.Context, tenantID uuid.UUID, slug, currency string) (*domain.Product, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(schema)

	var (
		p          domain.Product
		deviceType string
		features   []byte
		terms      []byte
	)
	err = r.db.QueryRowxContext(ctx, fmt.Sprintf(`
		SELECT id, slug, name, COALESCE(subtitle,''), COALESCE(description,''), device_type::text,
			subscription_days, codes_per_unit, features, terms, COALESCE(video_url,''), COALESCE(image_s3_key,''),
			is_published, purchase_count, rating_avg, rating_count, sort_order, created_at
		FROM %s.products
		WHERE slug = $1 AND is_published = true
	`, q), slug).Scan(&p.ID, &p.Slug, &p.Name, &p.Subtitle, &p.Description, &deviceType,
		&p.SubscriptionDays, &p.CodesPerUnit, &features, &terms, &p.VideoURL, &p.ImageURL,
		&p.IsPublished, &p.PurchaseCount, &p.RatingAvg, &p.RatingCount, &p.SortOrder, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	p.DeviceType = deviceType
	p.Features = decodeStringArray(features)
	p.Terms = decodeStringArray(terms)

	prices := []domain.ProductPrice{}
	if err := r.db.SelectContext(ctx, &prices, fmt.Sprintf(`
		SELECT currency, amount, compare_at_amount
		FROM %s.product_prices WHERE product_id = $1
		ORDER BY currency
	`, q), p.ID); err != nil {
		return nil, err
	}
	p.Prices = prices
	return &p, nil
}

// CreateOrder prices the cart from the catalog for the given currency (never trusting client prices),
// then inserts the order and its items in one transaction and returns the pending order.
func (r *shopRepo) CreateOrder(ctx context.Context, tenantID uuid.UUID, email, phone, currency string, lines []domain.CartLine) (*domain.Order, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := r.db.SetDedicatedSchema(ctx, tx, schema); err != nil {
		return nil, err
	}

	order := &domain.Order{Email: email, Phone: phone, Currency: currency, Status: "pending"}
	var subtotal float64

	for _, line := range lines {
		if line.Qty < 1 {
			return nil, fmt.Errorf("invalid quantity for %q", line.Slug)
		}
		var (
			productID uuid.UUID
			name      string
			amount    float64
		)
		err := tx.QueryRowxContext(ctx, `
			SELECT p.id, p.name, pr.amount
			FROM products p
			JOIN product_prices pr ON pr.product_id = p.id AND pr.currency = $2
			WHERE p.slug = $1 AND p.is_published = true
		`, line.Slug, currency).Scan(&productID, &name, &amount)
		if err == sql.ErrNoRows {
			return nil, errProductUnavailable
		}
		if err != nil {
			return nil, err
		}

		subtotal += amount * float64(line.Qty)
		order.Items = append(order.Items, domain.OrderItem{
			ProductID: productID, ProductName: name, Qty: line.Qty, UnitAmount: amount, Currency: currency,
		})
	}

	order.Subtotal = subtotal
	order.Total = subtotal

	if err := tx.QueryRowxContext(ctx, `
		INSERT INTO orders (email, phone, currency, subtotal, total, status)
		VALUES ($1, NULLIF($2,''), $3, $4, $5, 'pending')
		RETURNING id, created_at
	`, email, phone, currency, order.Subtotal, order.Total).Scan(&order.ID, &order.CreatedAt); err != nil {
		return nil, err
	}

	for i := range order.Items {
		it := &order.Items[i]
		if err := tx.QueryRowxContext(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, qty, unit_amount, currency)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, order.ID, it.ProductID, it.ProductName, it.Qty, it.UnitAmount, it.Currency).Scan(&it.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
}

func (r *shopRepo) GetOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(schema)

	var (
		o        domain.Order
		proofKey string
	)
	err = r.db.QueryRowxContext(ctx, fmt.Sprintf(`
		SELECT id, email, COALESCE(phone,''), currency, subtotal, total, status,
			COALESCE(provider,''), COALESCE(provider_ref,''), COALESCE(payment_proof_s3_key,''),
			created_at, paid_at, fulfilled_at
		FROM %s.orders WHERE id = $1
	`, q), orderID).Scan(&o.ID, &o.Email, &o.Phone, &o.Currency, &o.Subtotal, &o.Total, &o.Status,
		&o.Provider, &o.ProviderRef, &proofKey, &o.CreatedAt, &o.PaidAt, &o.FulfilledAt)
	if err != nil {
		return nil, err
	}
	o.HasPaymentProof = proofKey != ""

	if err := r.db.SelectContext(ctx, &o.Items, fmt.Sprintf(`
		SELECT id, product_id, product_name, qty, unit_amount, currency
		FROM %s.order_items WHERE order_id = $1 ORDER BY product_name
	`, q), orderID); err != nil {
		return nil, err
	}

	codes := []string{}
	if err := r.db.SelectContext(ctx, &codes, fmt.Sprintf(`
		SELECT code FROM %s.activation_codes WHERE order_id = $1 ORDER BY code
	`, q), orderID); err != nil {
		return nil, err
	}
	o.Codes = codes
	return &o, nil
}

// SetOrderProof stores the object key of the customer's uploaded transfer screenshot.
func (r *shopRepo) SetOrderProof(ctx context.Context, tenantID, orderID uuid.UUID, key string) error {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.orders SET payment_proof_s3_key = $2 WHERE id = $1
	`, postgres.QuoteIdentifier(schema)), orderID, key)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// OrderProofKey returns the stored proof key for an order ("" when none).
func (r *shopRepo) OrderProofKey(ctx context.Context, tenantID, orderID uuid.UUID) (string, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return "", err
	}
	var key string
	err = r.db.GetContext(ctx, &key, fmt.Sprintf(`
		SELECT COALESCE(payment_proof_s3_key,'') FROM %s.orders WHERE id = $1
	`, postgres.QuoteIdentifier(schema)), orderID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return key, err
}

func (r *shopRepo) AttachCheckout(ctx context.Context, tenantID, orderID uuid.UUID, provider, providerRef string) error {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.orders SET provider = $2, provider_ref = $3
		WHERE id = $1 AND status = 'pending'
	`, postgres.QuoteIdentifier(schema)), orderID, provider, providerRef)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("order %s not found or not pending", orderID)
	}
	return nil
}

// FulfillByProviderRef locks the order matching a payment-gateway reference and fulfills it if it has
// not already been. Used by the automated webhook path. Idempotent: a second call for an
// already-fulfilled order returns applied=false and changes nothing.
func (r *shopRepo) FulfillByProviderRef(ctx context.Context, tenantID uuid.UUID, provider, providerRef string) (*domain.Order, bool, error) {
	return r.fulfill(ctx, tenantID,
		`SELECT id, email, COALESCE(phone,''), currency, subtotal, total, status
		 FROM orders WHERE provider = $1 AND provider_ref = $2 FOR UPDATE`,
		[]interface{}{provider, providerRef},
		fmt.Sprintf("no order for %s ref %s", provider, providerRef), false)
}

// FulfillByID locks an order by id and fulfills it — used for MANUAL payment confirmation from the
// admin dashboard (there is no payment gateway). Marks the order fulfilled and mints the codes.
// Idempotent: confirming an already-fulfilled order returns applied=false and changes nothing.
func (r *shopRepo) FulfillByID(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, bool, error) {
	return r.fulfill(ctx, tenantID,
		`SELECT id, email, COALESCE(phone,''), currency, subtotal, total, status
		 FROM orders WHERE id = $1 FOR UPDATE`,
		[]interface{}{orderID},
		fmt.Sprintf("order %s not found", orderID), true)
}

// fulfill is the shared fulfillment core: it locks the order via lockSQL, and if not already
// fulfilled, mints one activation code per purchased unit (qty * codes_per_unit), bumps each
// product's purchase counter, and marks the order fulfilled. When manual is true the payment is
// attributed to "manual" (admin-confirmed) rather than a gateway.
func (r *shopRepo) fulfill(ctx context.Context, tenantID uuid.UUID, lockSQL string, lockArgs []interface{}, notFoundMsg string, manual bool) (*domain.Order, bool, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, false, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()
	if err := r.db.SetDedicatedSchema(ctx, tx, schema); err != nil {
		return nil, false, err
	}

	var order domain.Order
	err = tx.QueryRowxContext(ctx, lockSQL, lockArgs...).Scan(&order.ID, &order.Email, &order.Phone,
		&order.Currency, &order.Subtotal, &order.Total, &order.Status)
	if err == sql.ErrNoRows {
		return nil, false, fmt.Errorf("%s", notFoundMsg)
	}
	if err != nil {
		return nil, false, err
	}

	if order.Status == "fulfilled" {
		return &order, false, nil // already fulfilled — replayed webhook or double confirmation
	}

	// Items joined with their product so we know how many codes to mint and their device/expiry.
	type fulfillItem struct {
		ItemID           uuid.UUID `db:"item_id"`
		Qty              int       `db:"qty"`
		CodesPerUnit     int       `db:"codes_per_unit"`
		DeviceType       string    `db:"device_type"`
		SubscriptionDays int       `db:"subscription_days"`
	}
	var items []fulfillItem
	if err := tx.SelectContext(ctx, &items, `
		SELECT oi.id AS item_id, oi.qty, p.codes_per_unit, p.device_type::text AS device_type, p.subscription_days
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
	`, order.ID); err != nil {
		return nil, false, err
	}

	for _, it := range items {
		expiresAt := time.Now().Add(time.Duration(it.SubscriptionDays) * 24 * time.Hour)
		total := it.Qty * max(it.CodesPerUnit, 1)
		for minted := 0; minted < total; {
			code, err := generateActivationCode()
			if err != nil {
				return nil, false, err
			}
			res, err := tx.ExecContext(ctx, `
				INSERT INTO activation_codes (code, type, device_type, max_devices, max_uses, expires_at, order_id, order_item_id)
				VALUES ($1, 'time_bound'::activation_code_type, $2::allowed_device_type, 1, 1, $3, $4, $5)
				ON CONFLICT (code) DO NOTHING
			`, code, it.DeviceType, expiresAt, order.ID, it.ItemID)
			if err != nil {
				return nil, false, err
			}
			if n, _ := res.RowsAffected(); n == 1 {
				minted++
			}
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE products SET purchase_count = purchase_count + $2, updated_at = NOW()
			WHERE id = (SELECT product_id FROM order_items WHERE id = $1)
		`, it.ItemID, it.Qty); err != nil {
			return nil, false, err
		}
	}

	// For a manual confirmation, attribute the payment to "manual" when no gateway was recorded.
	statusSQL := `UPDATE orders SET status = 'fulfilled', paid_at = COALESCE(paid_at, NOW()), fulfilled_at = NOW() WHERE id = $1`
	if manual {
		statusSQL = `UPDATE orders SET status = 'fulfilled', provider = COALESCE(NULLIF(provider,''),'manual'), paid_at = COALESCE(paid_at, NOW()), fulfilled_at = NOW() WHERE id = $1`
	}
	if _, err := tx.ExecContext(ctx, statusSQL, order.ID); err != nil {
		return nil, false, err
	}

	// Read back the minted codes for the caller.
	codes := []string{}
	if err := tx.SelectContext(ctx, &codes, `SELECT code FROM activation_codes WHERE order_id = $1 ORDER BY code`, order.ID); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	order.Status = "fulfilled"
	order.Codes = codes
	return &order, true, nil
}

// SubmitProductRating records a 1-5 rating for a product (by slug). One rating per IP per product per
// day; a repeat from the same IP that day updates it. The products.rating_avg/count are refreshed by
// a DB trigger.
func (r *shopRepo) SubmitProductRating(ctx context.Context, tenantID uuid.UUID, slug string, userID *string, rating int, comment *string, ipHash string) error {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}
	q := postgres.QuoteIdentifier(schema)

	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]s.product_ratings (product_id, user_id, rating, comment, ip_hash)
		SELECT p.id, NULLIF($2,'')::uuid, $3, NULLIF($4,''), $5
		FROM %[1]s.products p WHERE p.slug = $1 AND p.is_published = true
		ON CONFLICT (product_id, ip_hash, rating_date)
		DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment, updated_at = NOW()
	`, q), slug, derefOrEmpty(userID), rating, derefOrEmpty(comment), ipHash)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows // no such published product
	}
	return nil
}

// ListProductReviews returns recent non-empty reviews for a product (by slug).
func (r *shopRepo) ListProductReviews(ctx context.Context, tenantID uuid.UUID, slug string, limit int) ([]domain.ProductReview, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	reviews := []domain.ProductReview{}
	if err := r.db.SelectContext(ctx, &reviews, fmt.Sprintf(`
		SELECT pr.rating, COALESCE(pr.comment,'') AS comment, pr.created_at::text AS created_at
		FROM %[1]s.product_ratings pr
		JOIN %[1]s.products p ON p.id = pr.product_id
		WHERE p.slug = $1 AND pr.comment IS NOT NULL AND pr.comment <> ''
		ORDER BY pr.created_at DESC
		LIMIT $2
	`, postgres.QuoteIdentifier(schema)), slug, limit); err != nil {
		return nil, err
	}
	return reviews, nil
}

// ProductImageKey returns the S3 key of a published product's image, or "" if none.
func (r *shopRepo) ProductImageKey(ctx context.Context, tenantID uuid.UUID, slug string) (string, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return "", err
	}
	var key string
	err = r.db.GetContext(ctx, &key, fmt.Sprintf(`
		SELECT COALESCE(image_s3_key,'') FROM %s.products WHERE slug = $1 AND is_published = true
	`, postgres.QuoteIdentifier(schema)), slug)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return key, err
}

// ListOrdersByEmail returns the order history for a customer email (newest first), for order tracking.
func (r *shopRepo) ListOrdersByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]domain.OrderSummary, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	items := []domain.OrderSummary{}
	if err := r.db.SelectContext(ctx, &items, fmt.Sprintf(`
		SELECT o.id::text AS id, o.email, COALESCE(o.phone,'') AS phone, o.currency, o.total,
			o.status, COALESCE(o.provider,'') AS provider, o.created_at::text AS created_at,
			(SELECT COUNT(*) FROM %[1]s.activation_codes ac WHERE ac.order_id = o.id) AS code_count
		FROM %[1]s.orders o
		WHERE lower(o.email) = lower($1)
		ORDER BY o.created_at DESC
		LIMIT 100
	`, postgres.QuoteIdentifier(schema)), email); err != nil {
		return nil, err
	}
	return items, nil
}

// CodeStatus returns the read-only status of an activation code WITHOUT consuming a device slot, for
// the storefront activation page. Returns nil (not an error) when the code does not exist.
func (r *shopRepo) CodeStatus(ctx context.Context, tenantID uuid.UUID, code string) (*domain.CodeStatus, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	var row struct {
		domain.CodeStatus
		Expired bool `db:"expired"`
	}
	err = r.db.GetContext(ctx, &row, fmt.Sprintf(`
		SELECT device_type::text AS device_type, max_devices, current_device_count, max_uses,
			current_uses, is_revoked, expires_at::text AS expires_at, first_used_at::text AS first_used_at,
			(expires_at IS NOT NULL AND expires_at <= NOW()) AS expired
		FROM %s.activation_codes WHERE code = $1
	`, postgres.QuoteIdentifier(schema)), strings.TrimSpace(code))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	st := row.CodeStatus
	st.Expired = row.Expired
	st.Valid = !st.IsRevoked && !st.Expired && st.CurrentUses < st.MaxUses
	return &st, nil
}

// InstallableApp returns the newest published app that has a signed version (the storefront's OTA
// install target). Returns nil (no error) when nothing is installable yet.
func (r *shopRepo) InstallableApp(ctx context.Context, tenantID uuid.UUID) (*domain.InstallableApp, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(schema)
	var app domain.InstallableApp
	err = r.db.GetContext(ctx, &app, fmt.Sprintf(`
		SELECT a.name AS name, v.id::text AS version_id, v.version AS version
		FROM %[1]s.applications a
		JOIN %[1]s.application_versions v ON v.application_id = a.id
		WHERE a.is_published = true AND v.signing_status = 'signed' AND v.signed_ipa_s3_key IS NOT NULL
		ORDER BY v.created_at DESC
		LIMIT 1
	`, q))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *shopRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		return "", err
	}
	return schemaName, nil
}

// scanProductRow scans a product joined with a single price row (list query shape).
func scanProductRow(rows *sqlx.Rows) (*domain.Product, domain.ProductPrice, error) {
	var (
		p          domain.Product
		deviceType string
		features   []byte
		terms      []byte
		price      domain.ProductPrice
	)
	if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Subtitle, &p.Description, &deviceType,
		&p.SubscriptionDays, &p.CodesPerUnit, &features, &terms, &p.VideoURL, &p.ImageURL,
		&p.IsPublished, &p.PurchaseCount, &p.RatingAvg, &p.RatingCount, &p.SortOrder, &p.CreatedAt,
		&price.Currency, &price.Amount, &price.CompareAt); err != nil {
		return nil, domain.ProductPrice{}, err
	}
	p.DeviceType = deviceType
	p.Features = decodeStringArray(features)
	p.Terms = decodeStringArray(terms)
	return &p, price, nil
}

// decodeStringArray parses a JSONB string-array column, tolerating NULL/empty as an empty slice.
func decodeStringArray(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	if out == nil {
		return []string{}
	}
	// Drop blank entries that can creep in from editing.
	cleaned := out[:0]
	for _, s := range out {
		if strings.TrimSpace(s) != "" {
			cleaned = append(cleaned, s)
		}
	}
	return cleaned
}
