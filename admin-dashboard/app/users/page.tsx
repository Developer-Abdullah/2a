import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminUser } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

export default async function UsersPage() {
  const t = await getT();
  let items: AdminUser[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AdminUser[] }>("/v1/admin/users");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  return (
    <DashboardShell title="title.users">
      <p className="mb-4 text-sm text-muted-foreground">{items.length} {t("usr.count")}</p>
      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("usr.no_users")}</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("usr.col.name")}</TableHead>
                  <TableHead>{t("usr.col.devices")}</TableHead>
                  <TableHead>{t("usr.col.joined")}</TableHead>
                  <TableHead>{t("usr.col.last_seen")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell className="font-medium text-foreground">
                      {u.display_name || <span className="text-muted-foreground">{t("common.unknown")}</span>}
                    </TableCell>
                    <TableCell className="text-foreground/80">{u.device_count}</TableCell>
                    <TableCell className="text-muted-foreground">{u.created_at.slice(0, 10)}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {u.last_seen_at ? u.last_seen_at.slice(0, 10) : <span className="text-muted-foreground">—</span>}
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
