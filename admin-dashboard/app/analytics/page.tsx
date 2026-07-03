import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminStats, RatingsResponse } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export const dynamic = "force-dynamic";

export default async function AnalyticsPage() {
  let stats: AdminStats | null = null;
  let ratings: RatingsResponse | null = null;
  let error = "";
  try {
    stats = await adminGet<AdminStats>("/v1/admin/stats/overview");
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load stats";
  }
  try {
    ratings = await adminGet<RatingsResponse>("/v1/admin/ratings");
  } catch {
    ratings = null;
  }

  const cards = stats
    ? [
        { label: "Applications", value: stats.apps },
        { label: "Versions", value: stats.versions },
        { label: "Users", value: stats.users },
        { label: "Active devices", value: stats.devices },
        { label: "Activation codes", value: stats.codes },
        { label: "Notifications", value: stats.notifications },
        { label: "Avg rating", value: stats.avg_rating.toFixed(2) },
        { label: "Total ratings", value: stats.rating_count },
      ]
    : [];
  const max = ratings ? Math.max(1, ...Object.values(ratings.summary.breakdown)) : 1;

  return (
    <DashboardShell title="title.analytics">
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
            {cards.map((c) => (
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

          {ratings ? (
            <Card>
              <CardHeader>
                <CardTitle>Rating distribution</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {[5, 4, 3, 2, 1].map((n) => {
                  const c = ratings.summary.breakdown[String(n)] ?? 0;
                  return (
                    <div key={n} className="flex items-center gap-3">
                      <span className="w-10 text-sm text-slate-600">{n} ★</span>
                      <div className="h-2 flex-1 rounded bg-slate-100">
                        <div className="h-2 rounded bg-blue-500" style={{ width: `${(c / max) * 100}%` }} />
                      </div>
                      <span className="w-8 text-right text-sm text-slate-500">{c}</span>
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
