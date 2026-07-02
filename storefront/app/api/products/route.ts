import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Client-accessible product list (proxied with the tenant header) so the cart can price its lines
// for the active currency without exposing the backend or its tenant header to the browser.
export async function GET(req: NextRequest) {
  const currency = req.nextUrl.searchParams.get("currency") || "EGP";
  const res = await fetch(`${BACKEND}/shop/products?currency=${encodeURIComponent(currency)}`, {
    headers: { "X-Tenant-Slug": STORE_SLUG },
    cache: "no-store",
  });
  const body = await res.json().catch(() => null);
  return NextResponse.json({ products: body?.data?.products ?? [] }, { status: res.ok ? 200 : 502 });
}
