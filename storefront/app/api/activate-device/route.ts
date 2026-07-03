import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Binds an activation code to a device (consumes a device/use slot) via the core API's activation
// validate endpoint. The device_id is a stable per-browser fingerprint generated on the client.
export async function POST(req: NextRequest) {
  let body: { code?: string; device_id?: string; device_type?: string };
  try {
    body = await req.json();
  } catch {
    return NextResponse.json({ error: "invalid_request" }, { status: 400 });
  }
  const { code, device_id, device_type } = body;
  if (!code || !device_id) {
    return NextResponse.json({ error: "missing_fields" }, { status: 400 });
  }

  const res = await fetch(`${BACKEND}/v1/activation/validate`, {
    method: "POST",
    headers: { "X-Tenant-Slug": STORE_SLUG, "Content-Type": "application/json" },
    body: JSON.stringify({ code, device_id, device_type: device_type || "iphone" }),
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  if (!res.ok || !data?.success) {
    return NextResponse.json({ error: data?.error?.code || "activation_failed" }, { status: res.status || 400 });
  }
  return NextResponse.json({ ok: true });
}
