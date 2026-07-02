-- Storefront catalog. A product is a sellable SKU (e.g. a one-year "Plus" subscription code). This
-- is distinct from `applications` (the IPAs the platform signs/distributes): a product is what the
-- customer pays for, and fulfilling a paid order mints an activation_code (see 022).

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Short URL slug used on the storefront, e.g. /PdDWGXW. Unique, stable, human-shareable.
    slug VARCHAR(64) NOT NULL UNIQUE,

    name VARCHAR(255) NOT NULL,
    subtitle VARCHAR(255),
    description TEXT,

    -- Which hardware the minted code activates. Reuses the enum defined in 003_create_activation_codes.
    device_type allowed_device_type NOT NULL DEFAULT 'both',

    -- Subscription length the minted activation code is valid for.
    subscription_days INT NOT NULL DEFAULT 365 CHECK (subscription_days > 0),

    -- How many activation codes one purchased unit yields (usually 1).
    codes_per_unit INT NOT NULL DEFAULT 1 CHECK (codes_per_unit >= 1),

    -- Marketing content: bullet features and terms are ordered JSON string arrays.
    features JSONB NOT NULL DEFAULT '[]',
    terms JSONB NOT NULL DEFAULT '[]',

    -- Optional how-to-activate tutorial video and product image.
    video_url VARCHAR(512),
    image_s3_key VARCHAR(512),

    is_published BOOLEAN NOT NULL DEFAULT false,

    -- "تم شراؤه X مرة" counter, incremented on each fulfilled order unit.
    purchase_count INT NOT NULL DEFAULT 0 CHECK (purchase_count >= 0),

    -- Denormalized rating aggregate shown on the product card (source of truth stays in ratings).
    rating_avg NUMERIC(3, 2) NOT NULL DEFAULT 0 CHECK (rating_avg >= 0 AND rating_avg <= 5),
    rating_count INT NOT NULL DEFAULT 0 CHECK (rating_count >= 0),

    -- Manual ordering on the storefront grid (lower = earlier).
    sort_order INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_published ON products (is_published, sort_order);
