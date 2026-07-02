package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProductPrice is one currency's price for a product. CompareAt is the optional struck-through
// "was" price used to convey a discount on the storefront.
type ProductPrice struct {
	Currency  string   `db:"currency" json:"currency"`
	Amount    float64  `db:"amount" json:"amount"`
	CompareAt *float64 `db:"compare_at_amount" json:"compare_at,omitempty"`
}

// Product is a sellable storefront SKU (e.g. a one-year subscription code). Prices holds one entry
// per supported currency (EGP, KWD).
type Product struct {
	ID               uuid.UUID      `db:"id" json:"id"`
	Slug             string         `db:"slug" json:"slug"`
	Name             string         `db:"name" json:"name"`
	Subtitle         string         `db:"subtitle" json:"subtitle"`
	Description      string         `db:"description" json:"description"`
	DeviceType       string         `db:"device_type" json:"device_type"`
	SubscriptionDays int            `db:"subscription_days" json:"subscription_days"`
	CodesPerUnit     int            `db:"codes_per_unit" json:"codes_per_unit"`
	Features         []string       `db:"-" json:"features"`
	Terms            []string       `db:"-" json:"terms"`
	VideoURL         string         `db:"video_url" json:"video_url"`
	ImageURL         string         `db:"image_url" json:"image_url"`
	IsPublished      bool           `db:"is_published" json:"is_published"`
	PurchaseCount    int            `db:"purchase_count" json:"purchase_count"`
	RatingAvg        float64        `db:"rating_avg" json:"rating_avg"`
	RatingCount      int            `db:"rating_count" json:"rating_count"`
	SortOrder        int            `db:"sort_order" json:"sort_order"`
	Prices           []ProductPrice `db:"-" json:"prices"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
}

// OrderItem is one purchased line. Name and unit price are snapshotted at purchase time.
type OrderItem struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ProductID   uuid.UUID `db:"product_id" json:"product_id"`
	ProductName string    `db:"product_name" json:"product_name"`
	Qty         int       `db:"qty" json:"qty"`
	UnitAmount  float64   `db:"unit_amount" json:"unit_amount"`
	Currency    string    `db:"currency" json:"currency"`
}

// Order is a customer checkout. Codes is populated once the order is fulfilled.
type Order struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Email       string      `db:"email" json:"email"`
	Phone       string      `db:"phone" json:"phone"`
	Currency    string      `db:"currency" json:"currency"`
	Subtotal    float64     `db:"subtotal" json:"subtotal"`
	Total       float64     `db:"total" json:"total"`
	Status      string      `db:"status" json:"status"`
	Provider    string      `db:"provider" json:"provider"`
	ProviderRef string      `db:"provider_ref" json:"provider_ref"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	PaidAt      *time.Time  `db:"paid_at" json:"paid_at,omitempty"`
	FulfilledAt *time.Time  `db:"fulfilled_at" json:"fulfilled_at,omitempty"`
	Items       []OrderItem `db:"-" json:"items"`
	Codes       []string    `db:"-" json:"codes"`
}

// CartLine is one requested product+quantity in a checkout request.
type CartLine struct {
	Slug string `json:"slug" binding:"required"`
	Qty  int    `json:"qty" binding:"required"`
}

// PriceInput is one currency's price when an admin creates/updates a product.
type PriceInput struct {
	Currency  string   `json:"currency" binding:"required"`
	Amount    float64  `json:"amount"`
	CompareAt *float64 `json:"compare_at"`
}

// ProductInput is the admin-editable shape of a product. Derived fields (purchase_count, rating_*)
// are not set here — they are maintained by the system.
type ProductInput struct {
	Slug             string       `json:"slug"`
	Name             string       `json:"name"`
	Subtitle         string       `json:"subtitle"`
	Description      string       `json:"description"`
	DeviceType       string       `json:"device_type"`
	SubscriptionDays int          `json:"subscription_days"`
	CodesPerUnit     int          `json:"codes_per_unit"`
	Features         []string     `json:"features"`
	Terms            []string     `json:"terms"`
	VideoURL         string       `json:"video_url"`
	IsPublished      bool         `json:"is_published"`
	SortOrder        int          `json:"sort_order"`
	Prices           []PriceInput `json:"prices"`
}

// OrderSummary is the admin list row for an order.
type OrderSummary struct {
	ID        string  `db:"id" json:"id"`
	Email     string  `db:"email" json:"email"`
	Phone     string  `db:"phone" json:"phone"`
	Currency  string  `db:"currency" json:"currency"`
	Total     float64 `db:"total" json:"total"`
	Status    string  `db:"status" json:"status"`
	Provider  string  `db:"provider" json:"provider"`
	CreatedAt string  `db:"created_at" json:"created_at"`
	CodeCount int     `db:"code_count" json:"code_count"`
}
