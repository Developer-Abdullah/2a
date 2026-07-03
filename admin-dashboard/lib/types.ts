// Shapes returned by the Admin API (:8081). Each endpoint wraps payloads in { success, data }.

export interface AppItem {
  id: string;
  name: string;
  category: string;
  icon_url: string;
  bundle_identifier: string;
  is_published: boolean;
  created_at: string;
}

export interface AppVersion {
  id: string;
  version: string;
  build_number: string;
  signing_status: string;
  signed_ipa_s3_key: string;
  size_bytes: number;
  created_at: string;
}

export interface RatingSummary {
  average: number;
  count: number;
  breakdown: Record<string, number>;
}

export interface RatingItem {
  id: string;
  user_id?: string;
  rating: number;
  comment?: string;
  created_at: string;
  updated_at: string;
}

export interface RatingsResponse {
  summary: RatingSummary;
  items: RatingItem[];
}

export interface NotificationItem {
  id: string;
  title: string;
  body: string;
  type: string;
  sent_at: string;
}

export interface ActivationCode {
  id: string;
  code: string;
  type: string;
  device_type: string;
  max_devices: number;
  current_device_count: number;
  max_uses: number;
  current_uses: number;
  is_revoked: boolean;
}

export interface AdminUser {
  id: string;
  display_name: string;
  device_count: number;
  created_at: string;
  last_seen_at: string;
}

export interface AdminDevice {
  id: string;
  user_id: string;
  device_type: string;
  model: string;
  enrollment_method: string;
  is_revoked: boolean;
  last_seen_at: string;
}

export interface AdminStats {
  apps: number;
  versions: number;
  users: number;
  devices: number;
  codes: number;
  notifications: number;
  avg_rating: number;
  rating_count: number;
}

export interface SalesStats {
  revenue_egp: number;
  revenue_kwd: number;
  orders_today: number;
  orders_month: number;
  pending_orders: number;
  fulfilled_orders: number;
  codes_issued: number;
  top_product_name: string;
  top_product_count: number;
}

export interface SigningJob {
  id: string;
  app_name: string;
  version: string;
  status: string;
  created_at: string;
}

export interface Certificate {
  id: string;
  label: string;
  is_active: boolean;
  added_at: string;
  expires_at: string;
}

export interface AuditEntry {
  id: string;
  action: string;
  resource_type: string;
  resource_id: string;
  ip_address: string;
  created_at: string;
}

export interface AdminAccount {
  id: string;
  email: string;
  role: string;
  created_at: string;
}

export interface UserDetail {
  user: AdminUser;
  devices: AdminDevice[];
}

export interface TenantSummary {
  slug: string;
  name: string;
  schema_name: string;
  status: string;
  created_at: string;
  apps: number;
}

// Storefront catalog + orders shapes (from the admin /v1/admin/products and /v1/admin/orders).
export interface ProductPrice {
  currency: string;
  amount: number;
  compare_at?: number | null;
}

export interface Product {
  id: string;
  slug: string;
  name: string;
  subtitle: string;
  description: string;
  device_type: string;
  subscription_days: number;
  codes_per_unit: number;
  features: string[];
  terms: string[];
  video_url: string;
  image_url: string;
  is_published: boolean;
  purchase_count: number;
  rating_avg: number;
  rating_count: number;
  sort_order: number;
  prices: ProductPrice[];
}

export interface OrderSummary {
  id: string;
  email: string;
  phone: string;
  currency: string;
  total: number;
  status: string;
  provider: string;
  created_at: string;
  code_count: number;
}

export interface OrderItem {
  id: string;
  product_id: string;
  product_name: string;
  qty: number;
  unit_amount: number;
  currency: string;
}

export interface OrderDetail {
  id: string;
  email: string;
  phone: string;
  currency: string;
  subtotal: number;
  total: number;
  status: string;
  provider: string;
  provider_ref: string;
  created_at: string;
  paid_at?: string | null;
  fulfilled_at?: string | null;
  items: OrderItem[];
  codes: string[];
  has_payment_proof: boolean;
}
