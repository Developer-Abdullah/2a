import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Submits a product rating to the core API (adds the tenant header server-side).
export async function POST(req: NextRequest) {
  let body: { slug?: string; rating?: number; comment?: string };
  try {
    body = await req.json();
  } catch {
    return NextResponse.json({ error: "invalid_request" }, { status: 400 });
  }
  const { slug, rating, comment } = body;
  if (!slug || !rating) {
    return NextResponse.json({ error: "missing_fields" }, { status: 400 });
  }
  const res = await fetch(`${BACKEND}/shop/products/${encodeURIComponent(slug)}/ratings`, {
    method: "POST",
    headers: { "X-Tenant-Slug": STORE_SLUG, "Content-Type": "application/json" },
    body: JSON.stringify({ rating, comment: comment || null }),
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  if (!res.ok || !data?.success) {
    return NextResponse.json({ error: data?.error?.code || "rate_failed" }, { status: res.status || 400 });
  }
  return NextResponse.json({ ok: true });
}
