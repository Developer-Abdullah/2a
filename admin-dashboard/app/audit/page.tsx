import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AuditEntry } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function AuditLogPage() {
  let items: AuditEntry[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AuditEntry[] }>("/v1/admin/audit");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load audit log";
  }

  return (
    <DashboardShell title="Audit Log">
      <p className="mb-4 text-sm text-slate-500">Immutable record of administrative actions.</p>
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-slate-500">No audit events recorded yet.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Resource</TableHead>
                  <TableHead>IP</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((a) => (
                  <TableRow key={a.id}>
                    <TableCell className="text-slate-500">{a.created_at.slice(0, 19)}</TableCell>
                    <TableCell>
                      <Badge variant="default">{a.action}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-slate-600">
                      {a.resource_type}/{a.resource_id}
                    </TableCell>
                    <TableCell className="text-slate-400">{a.ip_address || "—"}</TableCell>
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
