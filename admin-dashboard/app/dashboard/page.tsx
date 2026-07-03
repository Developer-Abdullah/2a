import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminStats, SalesStats } from "@/lib/types";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

function money(n: number, currency: string) {
  return `${n.toLocaleString(undefined, { maximumFractionDigits: currency === "KWD" ? 3 : 2 })} ${currency}`;
}

export default async function DashboardPage() {
  const t = await getT();
  let stats: AdminStats | null = null;
  let sales: SalesStats | null = null;
  let error = "";
  try {
    [stats, sales] = await Promise.all([
      adminGet<AdminStats>("/v1/admin/stats/overview"),
      adminGet<SalesStats>("/v1/admin/stats/sales"),
    ]);
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  const salesCards = sales
    ? [
        { label: t("dash.revenue_egp"), value: money(sales.revenue_egp, "EGP"), accent: "text-emerald-600 dark:text-emerald-400" },
        { label: t("dash.revenue_kwd"), value: money(sales.revenue_kwd, "KWD"), accent: "text-emerald-600 dark:text-emerald-400" },
        { label: t("dash.orders_today"), value: sales.orders_today.toLocaleString() },
        { label: t("dash.orders_month"), value: sales.orders_month.toLocaleString() },
        { label: t("dash.pending"), value: sales.pending_orders.toLocaleString(), accent: "text-amber-600 dark:text-amber-400" },
        { label: t("dash.fulfilled"), value: sales.fulfilled_orders.toLocaleString() },
        { label: t("dash.codes_issued"), value: sales.codes_issued.toLocaleString() },
        {
          label: t("dash.best_seller"),
          value: sales.top_product_name ? `${sales.top_product_name} (${sales.top_product_count})` : "—",
        },
      ]
    : [];

  const opsCards = stats
    ? [
        { label: t("dash.products"), value: stats.apps },
        { label: t("dash.customers"), value: stats.users },
        { label: t("dash.devices"), value: stats.devices },
        { label: t("dash.avg_rating"), value: stats.avg_rating.toFixed(2) },
      ]
    : [];

  return (
    <DashboardShell title="nav.dashboard">
      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : (
        <div className="space-y-8">
          <section>
            <h2 className="mb-3 text-sm font-semibold text-muted-foreground">{t("dash.sales")}</h2>
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
              {salesCards.map((c) => (
                <Card key={c.label}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">{c.label}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className={`text-2xl font-bold ${("accent" in c && c.accent) || "text-foreground"}`}>{c.value}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>

          <section>
            <h2 className="mb-3 text-sm font-semibold text-muted-foreground">{t("dash.store")}</h2>
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
              {opsCards.map((c) => (
                <Card key={c.label}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">{c.label}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-3xl font-bold text-foreground">{c.value}</p>
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
