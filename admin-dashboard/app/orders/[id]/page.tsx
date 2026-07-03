import Link from "next/link";
import { ArrowRight } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { OrderDetail } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import ConfirmOrderButton from "@/components/orders/ConfirmOrderButton";

export const dynamic = "force-dynamic";

const STATUS: Record<string, { label: string; variant: BadgeVariant }> = {
  pending: { label: "بانتظار الدفع", variant: "warning" },
  paid: { label: "مدفوع", variant: "default" },
  failed: { label: "فشل", variant: "danger" },
  fulfilled: { label: "مكتمل", variant: "success" },
};

export default async function OrderDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let order: OrderDetail | null = null;
  let error = "";
  try {
    const data = await adminGet<{ order: OrderDetail }>(`/v1/admin/orders/${id}`);
    order = data.order;
  } catch (e) {
    error = e instanceof Error ? e.message : "تعذّر تحميل الطلب";
  }

  const s = order ? STATUS[order.status] ?? { label: order.status, variant: "muted" as BadgeVariant } : null;

  return (
    <DashboardShell title="تفاصيل الطلب">
      <Link href="/orders" className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowRight className="ml-1 h-4 w-4" /> رجوع للطلبات
      </Link>

      {error || !order ? (
        <p className="text-sm text-red-600">{error || "الطلب غير موجود"}</p>
      ) : (
        <div dir="rtl" className="grid gap-6 text-right lg:grid-cols-3">
          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>الطلب #{order.id.slice(0, 8)}</span>
                {s ? <Badge variant={s.variant}>{s.label}</Badge> : null}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm text-slate-600">
              <div className="grid grid-cols-2 gap-3">
                <Info label="البريد" value={order.email} />
                <Info label="الجوال" value={order.phone || "—"} />
                <Info label="العملة" value={order.currency} />
                <Info label="بوابة الدفع" value={order.provider || "—"} />
                <Info label="التاريخ" value={new Date(order.created_at).toLocaleString("ar-EG")} />
                <Info label="الإجمالي" value={`${order.total} ${order.currency}`} />
              </div>

              <div>
                <h3 className="mb-2 font-semibold text-slate-900">العناصر</h3>
                <ul className="space-y-1">
                  {order.items.map((it) => (
                    <li key={it.id} className="flex items-center justify-between border-b border-slate-100 py-1">
                      <span>{it.product_name} × {it.qty}</span>
                      <span className="font-medium text-slate-800">{it.unit_amount * it.qty} {it.currency}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </CardContent>
          </Card>

          <Card className="h-fit">
            <CardHeader>
              <CardTitle>أكواد التفعيل</CardTitle>
            </CardHeader>
            <CardContent dir="rtl" className="space-y-4 text-right">
              {order.status !== "fulfilled" && (
                <div className="space-y-3 rounded-lg bg-amber-50 p-3">
                  {order.has_payment_proof ? (
                    <a href={`/api/order-proof/${order.id}`} target="_blank" rel="noopener noreferrer" className="block">
                      <div className="mb-1 text-xs font-medium text-emerald-700">إثبات التحويل المرفوع (اضغط للتكبير):</div>
                      <img src={`/api/order-proof/${order.id}`} alt="إثبات التحويل" className="max-h-56 w-full rounded-lg object-contain ring-1 ring-slate-200" />
                    </a>
                  ) : (
                    <p className="text-sm text-amber-800">لم يرفع العميل إثبات تحويل بعد.</p>
                  )}
                  <p className="text-sm text-amber-800">
                    بعد التحقق من الدفع، اضغط لتأكيد الطلب وإصدار الأكواد للعميل.
                  </p>
                  <ConfirmOrderButton id={order.id} />
                </div>
              )}

              {order.codes.length === 0 ? (
                order.status === "fulfilled" ? (
                  <p className="text-sm text-slate-500">لا توجد أكواد لهذا الطلب.</p>
                ) : null
              ) : (
                <ul className="space-y-2">
                  {order.codes.map((code) => (
                    <li key={code} className="rounded-lg bg-slate-100 px-3 py-2 font-mono text-sm font-bold tracking-wider text-slate-900">
                      {code}
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-xs text-slate-400">{label}</div>
      <div className="font-medium text-slate-800">{value}</div>
    </div>
  );
}
