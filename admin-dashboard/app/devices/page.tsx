import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminDevice } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function DevicesPage() {
  let items: AdminDevice[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AdminDevice[] }>("/v1/admin/devices");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load devices";
  }

  return (
    <DashboardShell title="title.devices">
      <p className="mb-4 text-sm text-slate-500">{items.length} device(s)</p>
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-slate-500">No devices yet.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Device</TableHead>
                  <TableHead>Enrollment</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last seen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((d) => (
                  <TableRow key={d.id}>
                    <TableCell>
                      <div className="font-medium capitalize text-slate-900">{d.device_type}</div>
                      <div className="text-xs text-slate-500">{d.model || "—"}</div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="muted">{d.enrollment_method}</Badge>
                    </TableCell>
                    <TableCell>
                      {d.is_revoked ? <Badge variant="danger">Revoked</Badge> : <Badge variant="success">Active</Badge>}
                    </TableCell>
                    <TableCell className="text-slate-500">
                      {d.last_seen_at ? d.last_seen_at.slice(0, 10) : "—"}
                    </TableCell>
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
