import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Looks up a customer's orders by email (adds the tenant header server-side).
export async function GET(req: NextRequest) {
  const email = req.nextUrl.searchParams.get("email") || "";
  if (!email) return NextResponse.json({ orders: [] });
  const res = await fetch(`${BACKEND}/shop/orders?email=${encodeURIComponent(email)}`, {
    headers: { "X-Tenant-Slug": STORE_SLUG },
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  return NextResponse.json({ orders: data?.data?.orders ?? [] }, { status: res.ok ? 200 : 502 });
}
