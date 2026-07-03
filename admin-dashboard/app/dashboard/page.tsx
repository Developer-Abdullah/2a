import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminStats, SalesStats } from "@/lib/types";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export const dynamic = "force-dynamic";

function money(n: number, currency: string) {
  return `${n.toLocaleString("ar-EG", { maximumFractionDigits: currency === "KWD" ? 3 : 2 })} ${currency}`;
}

export default async function DashboardPage() {
  let stats: AdminStats | null = null;
  let sales: SalesStats | null = null;
  let error = "";
  try {
    [stats, sales] = await Promise.all([
      adminGet<AdminStats>("/v1/admin/stats/overview"),
      adminGet<SalesStats>("/v1/admin/stats/sales"),
    ]);
  } catch (e) {
    error = e instanceof Error ? e.message : "تعذّر تحميل الإحصائيات";
  }

  const salesCards = sales
    ? [
        { label: "إيرادات (جنيه)", value: money(sales.revenue_egp, "EGP"), accent: "text-emerald-600" },
        { label: "إيرادات (دينار)", value: money(sales.revenue_kwd, "KWD"), accent: "text-emerald-600" },
        { label: "طلبات اليوم", value: sales.orders_today.toLocaleString("ar-EG") },
        { label: "طلبات هذا الشهر", value: sales.orders_month.toLocaleString("ar-EG") },
        { label: "بانتظار المراجعة", value: sales.pending_orders.toLocaleString("ar-EG"), accent: "text-amber-600" },
        { label: "طلبات مكتملة", value: sales.fulfilled_orders.toLocaleString("ar-EG") },
        { label: "أكواد صادرة", value: sales.codes_issued.toLocaleString("ar-EG") },
        {
          label: "الأكثر مبيعًا",
          value: sales.top_product_name ? `${sales.top_product_name} (${sales.top_product_count})` : "—",
        },
      ]
    : [];

  const opsCards = stats
    ? [
        { label: "المنتجات/التطبيقات", value: stats.apps },
        { label: "المستخدمون", value: stats.users },
        { label: "الأجهزة النشطة", value: stats.devices },
        { label: "متوسط التقييم", value: stats.avg_rating.toFixed(2) },
      ]
    : [];

  return (
    <DashboardShell title="لوحة التحكم">
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
        <div className="space-y-8" dir="rtl">
          <section>
            <h2 className="mb-3 text-sm font-semibold text-slate-500">المبيعات</h2>
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
              {salesCards.map((c) => (
                <Card key={c.label}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-slate-500">{c.label}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className={`text-2xl font-bold ${("accent" in c && c.accent) || "text-slate-900"}`}>{c.value}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>

          <section>
            <h2 className="mb-3 text-sm font-semibold text-slate-500">المتجر</h2>
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
              {opsCards.map((c) => (
                <Card key={c.label}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-slate-500">{c.label}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-3xl font-bold text-slate-900">{c.value}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>
        </div>
      )}
    </DashboardShell>
  );
}
