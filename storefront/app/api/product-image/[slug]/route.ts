import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Streams a product image from the core API (which reads it from object storage). The browser can't
// send the tenant header on an <img> request, so this same-origin proxy adds it.
export async function GET(_req: NextRequest, { params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const res = await fetch(`${BACKEND}/shop/products/${encodeURIComponent(slug)}/image`, {
    headers: { "X-Tenant-Slug": STORE_SLUG },
    cache: "no-store",
  });
  if (!res.ok || !res.body) {
    return new NextResponse(null, { status: 404 });
  }
  return new NextResponse(res.body, {
    status: 200,
    headers: {
      "Content-Type": res.headers.get("content-type") || "image/jpeg",
      "Cache-Control": "public, max-age=300",
    },
  });
}
