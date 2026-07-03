import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Checks an activation code's status via the core API (adds the tenant header server-side). Read-only:
// does not consume a device slot.
export async function GET(req: NextRequest) {
  const code = (req.nextUrl.searchParams.get("code") || "").trim();
  if (!code) return NextResponse.json({ error: "code_required" }, { status: 400 });

  const res = await fetch(`${BACKEND}/shop/activation?code=${encodeURIComponent(code)}`, {
    headers: { "X-Tenant-Slug": STORE_SLUG },
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  if (res.status === 404) return NextResponse.json({ error: "code_not_found" }, { status: 404 });
  if (!res.ok || !data?.success) {
    return NextResponse.json({ error: data?.error?.code || "lookup_failed" }, { status: res.status || 400 });
  }
  return NextResponse.json({ status: data.data.status });
}
