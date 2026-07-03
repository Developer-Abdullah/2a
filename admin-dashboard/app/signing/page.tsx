import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { SigningJob } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

function jobVariant(s: string): BadgeVariant {
  if (s === "completed") return "success";
  if (s === "failed") return "danger";
  if (s === "processing") return "default";
  return "warning";
}

export default async function SigningQueuePage() {
  let items: SigningJob[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: SigningJob[] }>("/v1/admin/signing/jobs");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load signing jobs";
  }

  return (
    <DashboardShell title="title.signing_queue">
      <p className="mb-4 text-sm text-slate-500">{items.length} job(s)</p>
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-slate-500">All signing queues are currently empty.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>App</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Queued</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((j) => (
                  <TableRow key={j.id}>
                    <TableCell className="font-medium text-slate-900">{j.app_name}</TableCell>
                    <TableCell className="text-slate-500">{j.version}</TableCell>
                    <TableCell>
                      <Badge variant={jobVariant(j.status)}>{j.status}</Badge>
                    </TableCell>
                    <TableCell className="text-slate-500">{j.created_at.slice(0, 16)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </DashboardShell>
  );
}
