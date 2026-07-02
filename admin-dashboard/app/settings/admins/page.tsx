import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminAccount } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

function roleVariant(role: string): BadgeVariant {
  return role === "super_admin" ? "default" : "muted";
}

export default async function AdminsPage() {
  let items: AdminAccount[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AdminAccount[] }>("/v1/admin/admins");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load team members";
  }

  return (
    <DashboardShell title="Team Members">
      <p className="mb-4 text-sm text-slate-500">{items.length} platform administrator(s)</p>
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-slate-500">No administrators found.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Email</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((a) => (
                  <TableRow key={a.id}>
                    <TableCell className="font-medium text-slate-900">{a.email}</TableCell>
                    <TableCell>
                      <Badge variant={roleVariant(a.role)}>{a.role}</Badge>
                    </TableCell>
                    <TableCell className="text-slate-500">{a.created_at.slice(0, 10)}</TableCell>
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
