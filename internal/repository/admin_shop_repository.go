package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

// adminShopRepo backs the store-admin catalog + orders views. Unlike shopRepo (public storefront), it
// returns unpublished products and never filters by currency, since the admin manages everything.
type adminShopRepo struct {
	db *postgres.DB
}

func NewAdminShopRepository(db *postgres.DB) *adminShopRepo {
	return &adminShopRepo{db: db}
}

// ListProducts returns every product (published or draft) with all its currency prices.
func (r *adminShopRepo) ListProducts(ctx context.Context, tenantID uuid.UUID) ([]domain.Product, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(schema)

	var rows []struct {
		domain.Product
		Features []byte `db:"features"`
		Terms    []byte `db:"terms"`
	}
	if err := r.db.SelectContext(ctx, &rows, fmt.Sprintf(`
		SELECT id, slug, name, COALESCE(subtitle,'') AS subtitle, COALESCE(description,'') AS description,
			device_type::text AS device_type, subscription_days, codes_per_unit,
			features, terms, COALESCE(video_url,'') AS video_url, COALESCE(image_s3_key,'') AS image_url,
			is_published, purchase_count, rating_avg, rating_count, sort_order, created_at
		FROM %s.products ORDER BY sort_order ASC, created_at DESC
	`, q)); err != nil {
		return nil, err
	}

	products := make([]domain.Product, 0, len(rows))
	ids := make([]string, 0, len(rows))
	for i := range rows {
		p := rows[i].Product
		p.Features = decodeStringArray(rows[i].Features)
		p.Terms = decodeStringArray(rows[i].Terms)
		p.Prices = []domain.ProductPrice{}
		products = append(products, p)
		ids = append(ids, p.ID.String())
	}
	if len(products) == 0 {
		return products, nil
	}

	// Attach prices in one pass.
	var prices []struct {
		ProductID uuid.UUID `db:"product_id"`
		domain.ProductPrice
	}
	if err := r.db.SelectContext(ctx, &prices, fmt.Sprintf(`
		SELECT product_id, currency, amount, compare_at_amount
		FROM %s.product_prices ORDER BY currency
	`, q)); err != nil {
		return nil, err
	}
	byID := map[uuid.UUID]int{}
	for i := range products {
		byID[products[i].ID] = i
	}
	for _, pr := range prices {
		if idx, ok := byID[pr.ProductID]; ok {
			products[idx].Prices = append(products[idx].Prices, pr.ProductPrice)
		}
	}
	return products, nil
}

// GetProduct returns a single product (any publish state) with all prices, by id.
func (r *adminShopRepo) GetProduct(ctx context.Context, tenantID uuid.UUID, id string) (*domain.Product, error) {
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
		FROM %s.products WHERE id = $1
	`, q), id).Scan(&p.ID, &p.Slug, &p.Name, &p.Subtitle, &p.Description, &deviceType,
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
		SELECT currency, amount, compare_at_amount FROM %s.product_prices WHERE product_id = $1 ORDER BY currency
	`, q), p.ID); err != nil {
		return nil, err
	}
	p.Prices = prices
	return &p, nil
}

// CreateProduct inserts a product and its price rows in one transaction, returning the new id.
func (r *adminShopRepo) CreateProduct(ctx context.Context, tenantID uuid.UUID, in domain.ProductInput) (string, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return "", err
	}
	q := postgres.QuoteIdentifier(schema)
	deviceType := normalizeProductDeviceType(in.DeviceType)
	features := marshalJSONArray(in.Features)
	terms := marshalJSONArray(in.Terms)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if err := r.db.SetDedicatedSchema(ctx, tx, schema); err != nil {
		return "", err
	}

	var id string
	if err := tx.QueryRowxContext(ctx, `
		INSERT INTO products
			(slug, name, subtitle, description, device_type, subscription_days, codes_per_unit,
			 features, terms, video_url, is_published, sort_order)
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), $5::allowed_device_type, $6, $7,
			$8::jsonb, $9::jsonb, NULLIF($10,''), $11, $12)
		RETURNING id::text
	`, in.Slug, in.Name, in.Subtitle, in.Description, deviceType, normalizeInt(in.SubscriptionDays, 365),
		normalizeInt(in.CodesPerUnit, 1), features, terms, in.VideoURL, in.IsPublished, in.SortOrder).Scan(&id); err != nil {
		return "", err
	}

	if err := upsertPricesTx(ctx, tx, q, id, in.Prices); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

// UpdateProduct updates the product core fields and replaces its price set.
func (r *adminShopRepo) UpdateProduct(ctx context.Context, tenantID uuid.UUID, id string, in domain.ProductInput) error {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}
	q := postgres.QuoteIdentifier(schema)
	deviceType := normalizeProductDeviceType(in.DeviceType)
	features := marshalJSONArray(in.Features)
	terms := marshalJSONArray(in.Terms)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := r.db.SetDedicatedSchema(ctx, tx, schema); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE products SET
			slug = $2, name = $3, subtitle = NULLIF($4,''), description = NULLIF($5,''),
			device_type = $6::allowed_device_type, subscription_days = $7, codes_per_unit = $8,
			features = $9::jsonb, terms = $10::jsonb, video_url = NULLIF($11,''),
			is_published = $12, sort_order = $13, updated_at = NOW()
		WHERE id = $1
	`, id, in.Slug, in.Name, in.Subtitle, in.Description, deviceType, normalizeInt(in.SubscriptionDays, 365),
		normalizeInt(in.CodesPerUnit, 1), features, terms, in.VideoURL, in.IsPublished, in.SortOrder)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	// Replace the price set: delete then re-insert, so removing a currency actually removes it.
	if _, err := tx.ExecContext(ctx, `DELETE FROM product_prices WHERE product_id = $1`, id); err != nil {
		return err
	}
	if err := upsertPricesTx(ctx, tx, q, id, in.Prices); err != nil {
		return err
	}
	return tx.Commit()
}

// SetPublished toggles the storefront visibility of a product.
func (r *adminShopRepo) SetPublished(ctx context.Context, tenantID uuid.UUID, id string, published bool) error {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.products SET is_published = $2, updated_at = NOW() WHERE id = $1
	`, postgres.QuoteIdentifier(schema)), id, published)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListOrders returns order summaries, newest first, optionally filtered by status.
func (r *adminShopRepo) ListOrders(ctx context.Context, tenantID uuid.UUID, status string, limit int) ([]domain.OrderSummary, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	args := []interface{}{limit}
	where := ""
	if status != "" {
		args = append(args, status)
		where = "WHERE o.status = $2"
	}
	query := fmt.Sprintf(`
		SELECT o.id::text AS id, o.email, COALESCE(o.phone,'') AS phone, o.currency, o.total,
			o.status, COALESCE(o.provider,'') AS provider, o.created_at::text AS created_at,
			(SELECT COUNT(*) FROM %[1]s.activation_codes ac WHERE ac.order_id = o.id) AS code_count
		FROM %[1]s.orders o
		%[2]s
		ORDER BY o.created_at DESC
		LIMIT $1
	`, postgres.QuoteIdentifier(schema), where)

	var items []domain.OrderSummary
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.OrderSummary{}
	}
	return items, nil
}

// GetOrder returns a single order with its items and any minted codes (admin detail view).
func (r *adminShopRepo) GetOrder(ctx context.Context, tenantID uuid.UUID, orderID string) (*domain.Order, error) {
	schema, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(schema)

	var o domain.Order
	err = r.db.QueryRowxContext(ctx, fmt.Sprintf(`
		SELECT id, email, COALESCE(phone,''), currency, subtotal, total, status,
			COALESCE(provider,''), COALESCE(provider_ref,''), created_at, paid_at, fulfilled_at
		FROM %s.orders WHERE id = $1
	`, q), orderID).Scan(&o.ID, &o.Email, &o.Phone, &o.Currency, &o.Subtotal, &o.Total, &o.Status,
		&o.Provider, &o.ProviderRef, &o.CreatedAt, &o.PaidAt, &o.FulfilledAt)
	if err != nil {
		return nil, err
	}
	if err := r.db.SelectContext(ctx, &o.Items, fmt.Sprintf(`
		SELECT id, product_id, product_name, qty, unit_amount, currency
		FROM %s.order_items WHERE order_id = $1 ORDER BY product_name
	`, q), o.ID); err != nil {
		return nil, err
	}
	codes := []string{}
	if err := r.db.SelectContext(ctx, &codes, fmt.Sprintf(`
		SELECT code FROM %s.activation_codes WHERE order_id = $1 ORDER BY code
	`, q), o.ID); err != nil {
		return nil, err
	}
	o.Codes = codes
	return &o, nil
}

func (r *adminShopRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		return "", err
	}
	return schemaName, nil
}

// upsertPricesTx writes the given price rows for a product inside an open tx (schema already on the
// search_path). compare_at is stored NULL when absent or not above amount.
func upsertPricesTx(ctx context.Context, tx interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}, _ string, productID string, prices []domain.PriceInput) error {
	for _, p := range prices {
		if p.Currency == "" {
			continue
		}
		var compareAt interface{}
		if p.CompareAt != nil && *p.CompareAt > p.Amount {
			compareAt = *p.CompareAt
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO product_prices (product_id, currency, amount, compare_at_amount)
			VALUES ($1::uuid, $2, $3, $4)
			ON CONFLICT (product_id, currency) DO UPDATE SET amount = EXCLUDED.amount, compare_at_amount = EXCLUDED.compare_at_amount
		`, productID, p.Currency, p.Amount, compareAt); err != nil {
			return err
		}
	}
	return nil
}

func marshalJSONArray(items []string) string {
	if items == nil {
		items = []string{}
	}
	b, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func normalizeProductDeviceType(dt string) string {
	switch dt {
	case "iphone", "ipad", "both":
		return dt
	default:
		return "both"
	}
}

func normalizeInt(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}
