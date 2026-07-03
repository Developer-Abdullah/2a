import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminStats, RatingsResponse } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

export default async function AnalyticsPage() {
  const t = await getT();
  let stats: AdminStats | null = null;
  let ratings: RatingsResponse | null = null;
  let error = "";
  try {
    stats = await adminGet<AdminStats>("/v1/admin/stats/overview");
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }
  try {
    ratings = await adminGet<RatingsResponse>("/v1/admin/ratings");
  } catch {
    ratings = null;
  }

  const cards = stats
    ? [
        { label: t("ana.applications"), value: stats.apps },
        { label: t("ana.versions"), value: stats.versions },
        { label: t("ana.users"), value: stats.users },
        { label: t("ana.active_devices"), value: stats.devices },
        { label: t("ana.activation_codes"), value: stats.codes },
        { label: t("ana.notifications"), value: stats.notifications },
        { label: t("ana.avg_rating"), value: stats.avg_rating.toFixed(2) },
        { label: t("ana.total_ratings"), value: stats.rating_count },
      ]
    : [];
  const max = ratings ? Math.max(1, ...Object.values(ratings.summary.breakdown)) : 1;

  return (
    <DashboardShell title="title.analytics">
      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : (
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
            {cards.map((c) => (
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

          {ratings ? (
            <Card>
              <CardHeader>
                <CardTitle>{t("ana.distribution")}</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {[5, 4, 3, 2, 1].map((n) => {
                  const c = ratings.summary.breakdown[String(n)] ?? 0;
                  return (
                    <div key={n} className="flex items-center gap-3">
                      <span className="w-10 text-sm text-foreground/80">{n} ★</span>
                      <div className="h-2 flex-1 rounded bg-muted">
                        <div className="h-2 rounded bg-brand-500" style={{ width: `${(c / max) * 100}%` }} />
                      </div>
                      <span className="w-8 text-end text-sm text-muted-foreground">{c}</span>
                    </div>
                  );
                })}
              </CardContent>
            </Card>
          ) : null}
        </div>
      )}
    </DashboardShell>
  );
}
