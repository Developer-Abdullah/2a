export const CURRENCIES = ["EGP", "KWD"] as const;
export type Currency = (typeof CURRENCIES)[number];

export const CURRENCY_LABELS: Record<Currency, string> = {
  EGP: "ج.م",
  KWD: "د.ك",
};

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
  purchase_count: number;
  rating_avg: number;
  rating_count: number;
  prices: ProductPrice[];
}

export interface OrderItem {
  id: string;
  product_id: string;
  product_name: string;
  qty: number;
  unit_amount: number;
  currency: string;
}

export interface Order {
  id: string;
  email: string;
  phone: string;
  currency: string;
  subtotal: number;
  total: number;
  status: "pending" | "paid" | "failed" | "fulfilled";
  provider: string;
  provider_ref: string;
  created_at: string;
  paid_at?: string | null;
  fulfilled_at?: string | null;
  items: OrderItem[];
  codes: string[];
  has_payment_proof: boolean;
}

export interface CodeStatus {
  valid: boolean;
  device_type: string;
  max_devices: number;
  current_device_count: number;
  max_uses: number;
  current_uses: number;
  is_revoked: boolean;
  expired: boolean;
  expires_at?: string | null;
  first_used_at?: string | null;
}
