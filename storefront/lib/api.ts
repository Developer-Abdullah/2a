import { cookies } from "next/headers";
import { Currency, Order, Product } from "./types";

// Server-side base URL for the Go core API. The storefront always talks to a single store tenant,
// identified by a fixed slug header (no subdomain needed in dev).
const BACKEND = process.env.BACKEND_API_URL || "http://localhost:8080";
const STORE_SLUG = process.env.STORE_TENANT_SLUG || "store";

export async function getCurrency(): Promise<Currency> {
  const c = (await cookies()).get("currency")?.value;
  return c === "KWD" ? "KWD" : "EGP";
}

async function apiGet<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${BACKEND}${path}`, {
      headers: { "X-Tenant-Slug": STORE_SLUG },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const body = await res.json();
    return (body?.data ?? null) as T;
  } catch {
    return null;
  }
}

export async function fetchProducts(currency: string): Promise<Product[]> {
  const data = await apiGet<{ products: Product[] }>(`/shop/products?currency=${currency}`);
  return data?.products ?? [];
}

export async function fetchProduct(slug: string): Promise<Product | null> {
  const data = await apiGet<{ product: Product }>(`/shop/products/${encodeURIComponent(slug)}`);
  return data?.product ?? null;
}

export async function fetchOrder(id: string): Promise<Order | null> {
  const data = await apiGet<{ order: Order }>(`/shop/orders/${encodeURIComponent(id)}`);
  return data?.order ?? null;
}

export interface Review {
  rating: number;
  comment: string;
  created_at: string;
}

export async function fetchReviews(slug: string): Promise<Review[]> {
  const data = await apiGet<{ reviews: Review[] }>(`/shop/products/${encodeURIComponent(slug)}/reviews`);
  return data?.reviews ?? [];
}

export { BACKEND, STORE_SLUG };
