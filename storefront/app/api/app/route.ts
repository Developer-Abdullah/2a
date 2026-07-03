import { NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Returns the installable "container" app + OTA manifest URL for the storefront (adds tenant header).
export async function GET() {
  const res = await fetch(`${BACKEND}/shop/app`, {
    headers: { "X-Tenant-Slug": STORE_SLUG },
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  if (!res.ok || !data?.success) {
    return NextResponse.json({ ready: false }, { status: res.ok ? 200 : 502 });
  }
  return NextResponse.json(data.data);
}
