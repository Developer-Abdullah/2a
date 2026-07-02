import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminStats } from "@/lib/types";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

export const dynamic = "force-dynamic";

export default async function DashboardPage() {
  let stats: AdminStats | null = null;
  let error = "";
  try {
    stats = await adminGet<AdminStats>("/v1/admin/stats/overview");
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load stats";
  }

  const cards = stats
    ? [
        { label: "Applications", value: stats.apps },
        { label: "Versions", value: stats.versions },
        { label: "Users", value: stats.users },
        { label: "Active devices", value: stats.devices },
        { label: "Activation codes", value: stats.codes },
        { label: "Notifications", value: stats.notifications },
        { label: "Avg. rating", value: stats.avg_rating.toFixed(2) },
        { label: "Total ratings", value: stats.rating_count },
      ]
    : [];

  return (
    <DashboardShell title="Dashboard">
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
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
      )}
    </DashboardShell>
  );
}
