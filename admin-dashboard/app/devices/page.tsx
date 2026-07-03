import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminDevice } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

export default async function DevicesPage() {
  const t = await getT();
  let items: AdminDevice[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AdminDevice[] }>("/v1/admin/devices");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  return (
    <DashboardShell title="title.devices">
      <p className="mb-4 text-sm text-muted-foreground">{items.length} {t("dev.count")}</p>
      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("dev.no_devices")}</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("dev.col.device")}</TableHead>
                  <TableHead>{t("dev.col.enrollment")}</TableHead>
                  <TableHead>{t("common.status")}</TableHead>
                  <TableHead>{t("dev.col.last_seen")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((d) => (
                  <TableRow key={d.id}>
                    <TableCell>
                      <div className="font-medium capitalize text-foreground">{d.device_type}</div>
                      <div className="text-xs text-muted-foreground">{d.model || "—"}</div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="muted">{d.enrollment_method}</Badge>
                    </TableCell>
                    <TableCell>
                      {d.is_revoked ? <Badge variant="danger">{t("common.revoked")}</Badge> : <Badge variant="success">{t("common.active")}</Badge>}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
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
