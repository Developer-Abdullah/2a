import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminUser } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function UsersPage() {
  let items: AdminUser[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AdminUser[] }>("/v1/admin/users");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load users";
  }

  return (
    <DashboardShell title="Users">
      <p className="mb-4 text-sm text-slate-500">{items.length} user(s)</p>
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-slate-500">No users yet.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Devices</TableHead>
                  <TableHead>Joined</TableHead>
                  <TableHead>Last seen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell className="font-medium text-slate-900">
                      {u.display_name || <span className="text-slate-400">Unknown</span>}
                    </TableCell>
                    <TableCell className="text-slate-600">{u.device_count}</TableCell>
                    <TableCell className="text-slate-500">{u.created_at.slice(0, 10)}</TableCell>
                    <TableCell className="text-slate-500">
                      {u.last_seen_at ? u.last_seen_at.slice(0, 10) : <span className="text-slate-400">—</span>}
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
