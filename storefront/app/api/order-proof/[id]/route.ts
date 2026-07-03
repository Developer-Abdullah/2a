import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Forwards a payment-proof screenshot upload to the core API (adds the tenant header server-side).
export async function POST(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const form = await req.formData();
  const file = form.get("file");
  if (!(file instanceof Blob)) {
    return NextResponse.json({ error: "no file" }, { status: 400 });
  }
  const forward = new FormData();
  forward.append("file", file, (file as File).name || "proof");

  const res = await fetch(`${BACKEND}/shop/orders/${encodeURIComponent(id)}/proof`, {
    method: "POST",
    headers: { "X-Tenant-Slug": STORE_SLUG },
    body: forward,
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  if (!res.ok || !data?.success) {
    return NextResponse.json({ error: data?.error?.code || "upload_failed" }, { status: res.status || 400 });
  }
  return NextResponse.json({ ok: true });
}
