import { NextRequest, NextResponse } from "next/server";
import { BACKEND, STORE_SLUG } from "@/lib/api";

// Manual-payment checkout: create a pending order priced from the catalog and return its id. There is
// no payment gateway — the customer pays offline and an admin confirms the order from the dashboard,
// which mints and delivers the codes.
export async function POST(req: NextRequest) {
  let payload: { email?: string; phone?: string; currency?: string; items?: { slug: string; qty: number }[] };
  try {
    payload = await req.json();
  } catch {
    return NextResponse.json({ error: "invalid_request" }, { status: 400 });
  }

  const { email, phone, currency, items } = payload;
  if (!email || !currency || !items?.length) {
    return NextResponse.json({ error: "missing_fields" }, { status: 400 });
  }

  const orderRes = await fetch(`${BACKEND}/shop/orders`, {
    method: "POST",
    headers: { "X-Tenant-Slug": STORE_SLUG, "Content-Type": "application/json" },
    body: JSON.stringify({ email, phone: phone || "", currency, items }),
    cache: "no-store",
  });
  const orderBody = await orderRes.json().catch(() => null);
  if (!orderRes.ok || !orderBody?.data?.order?.id) {
    return NextResponse.json(
      { error: orderBody?.error?.code || "order_failed", message: orderBody?.error?.message },
      { status: 400 }
    );
  }

  return NextResponse.json({ orderId: orderBody.data.order.id as string });
}
