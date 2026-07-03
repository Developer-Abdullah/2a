import { Star } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { RatingsResponse } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function RatingsPage() {
  let data: RatingsResponse | null = null;
  let error = "";
  try {
    data = await adminGet<RatingsResponse>("/v1/admin/ratings");
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load ratings";
  }

  const summary = data?.summary;
  const items = data?.items ?? [];
  const max = summary ? Math.max(1, ...Object.values(summary.breakdown)) : 1;

  return (
    <DashboardShell title="title.ratings">
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-slate-500">Average</CardTitle>
              </CardHeader>
              <CardContent className="flex items-center gap-2">
                <Star className="h-6 w-6 fill-amber-400 text-amber-400" />
                <span className="text-3xl font-bold">{summary?.average.toFixed(2) ?? "0.00"}</span>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-slate-500">Total reviews</CardTitle>
              </CardHeader>
              <CardContent>
                <span className="text-3xl font-bold">{summary?.count ?? 0}</span>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Breakdown</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {[5, 4, 3, 2, 1].map((n) => {
                const c = summary?.breakdown[String(n)] ?? 0;
                return (
                  <div key={n} className="flex items-center gap-3">
                    <span className="w-10 text-sm text-slate-600">{n} ★</span>
                    <div className="h-2 flex-1 rounded bg-slate-100">
                      <div className="h-2 rounded bg-amber-400" style={{ width: `${(c / max) * 100}%` }} />
                    </div>
                    <span className="w-8 text-right text-sm text-slate-500">{c}</span>
                  </div>
                );
              })}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Recent reviews</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {items.length === 0 ? (
                <p className="px-6 pb-6 text-sm text-slate-500">No ratings yet.</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Rating</TableHead>
                      <TableHead>Comment</TableHead>
                      <TableHead>Date</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((r) => (
                      <TableRow key={r.id}>
                        <TableCell className="font-medium text-amber-500">
                          {"★".repeat(r.rating)}
                          <span className="text-slate-300">{"★".repeat(5 - r.rating)}</span>
                        </TableCell>
                        <TableCell className="text-slate-600">
                          {r.comment || <span className="text-slate-400">—</span>}
                        </TableCell>
                        <TableCell className="text-slate-500">{r.created_at.slice(0, 10)}</TableCell>
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
