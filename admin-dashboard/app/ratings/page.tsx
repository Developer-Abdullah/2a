import { Star } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { RatingsResponse } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

export default async function RatingsPage() {
  const t = await getT();
  let data: RatingsResponse | null = null;
  let error = "";
  try {
    data = await adminGet<RatingsResponse>("/v1/admin/ratings");
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  const summary = data?.summary;
  const items = data?.items ?? [];
  const max = summary ? Math.max(1, ...Object.values(summary.breakdown)) : 1;

  return (
    <DashboardShell title="title.ratings">
      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : (
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{t("rate.average")}</CardTitle>
              </CardHeader>
              <CardContent className="flex items-center gap-2">
                <Star className="h-6 w-6 fill-amber-400 text-amber-400" />
                <span className="text-3xl font-bold text-foreground">{summary?.average.toFixed(2) ?? "0.00"}</span>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{t("rate.total_reviews")}</CardTitle>
              </CardHeader>
              <CardContent>
                <span className="text-3xl font-bold text-foreground">{summary?.count ?? 0}</span>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>{t("rate.breakdown")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {[5, 4, 3, 2, 1].map((n) => {
                const c = summary?.breakdown[String(n)] ?? 0;
                return (
                  <div key={n} className="flex items-center gap-3">
                    <span className="w-10 text-sm text-foreground/80">{n} ★</span>
                    <div className="h-2 flex-1 rounded bg-muted">
                      <div className="h-2 rounded bg-amber-400" style={{ width: `${(c / max) * 100}%` }} />
                    </div>
                    <span className="w-8 text-end text-sm text-muted-foreground">{c}</span>
                  </div>
                );
              })}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t("rate.recent")}</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {items.length === 0 ? (
                <p className="px-6 pb-6 text-sm text-muted-foreground">{t("rate.no_ratings")}</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("rate.col.rating")}</TableHead>
                      <TableHead>{t("rate.col.comment")}</TableHead>
                      <TableHead>{t("rate.col.date")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((r) => (
                      <TableRow key={r.id}>
                        <TableCell className="font-medium text-amber-500">
                          {"★".repeat(r.rating)}
                          <span className="text-muted-foreground/40">{"★".repeat(5 - r.rating)}</span>
                        </TableCell>
                        <TableCell className="text-foreground/80">
                          {r.comment || <span className="text-muted-foreground">—</span>}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{r.created_at.slice(0, 10)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
