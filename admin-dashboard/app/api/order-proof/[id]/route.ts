import { NextRequest, NextResponse } from "next/server";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";

const ADMIN_BASE = process.env.BACKEND_API_URL || "http://localhost:8081";
const TENANT_SLUG = process.env.ADMIN_TENANT_SLUG || "store";

// Streams the customer's uploaded payment screenshot from the admin API (attaches the admin bearer +
// tenant header so the <img> in the order page can load it).
export async function GET(_req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const session = await getServerSession(authOptions);
  if (!session) return new NextResponse(null, { status: 401 });

  const res = await fetch(`${ADMIN_BASE}/v1/admin/orders/${encodeURIComponent(id)}/proof`, {
    headers: {
      Authorization: `Bearer ${session.accessToken ?? ""}`,
      "X-Tenant-Slug": session.user?.tenantSlug || TENANT_SLUG,
    },
    cache: "no-store",
  });
  if (!res.ok || !res.body) return new NextResponse(null, { status: 404 });
  return new NextResponse(res.body, {
    status: 200,
    headers: { "Content-Type": res.headers.get("content-type") || "image/jpeg", "Cache-Control": "private, max-age=60" },
  });
}
