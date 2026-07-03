import { NextRequest, NextResponse } from "next/server";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";

const ADMIN_BASE = process.env.BACKEND_API_URL || "http://localhost:8081";
const TENANT_SLUG = process.env.ADMIN_TENANT_SLUG || "store";

// Forwards a product-image multipart upload to the admin API, attaching the admin bearer token and
// tenant header server-side (the browser has neither). Returns the stored object key.
export async function POST(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const session = await getServerSession(authOptions);
  if (!session) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const form = await req.formData();
  const file = form.get("file");
  if (!(file instanceof Blob)) {
    return NextResponse.json({ error: "no file" }, { status: 400 });
  }

  const forward = new FormData();
  forward.append("file", file, (file as File).name || "image");

  const res = await fetch(`${ADMIN_BASE}/v1/admin/products/${encodeURIComponent(id)}/image`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${session.accessToken ?? ""}`,
      "X-Tenant-Slug": session.user?.tenantSlug || TENANT_SLUG,
    },
    body: forward,
    cache: "no-store",
  });
  const data = await res.json().catch(() => null);
  if (!res.ok || !data?.success) {
    return NextResponse.json({ error: data?.error?.message || "upload failed" }, { status: res.status || 400 });
  }
  return NextResponse.json({ key: data.data.key });
}
