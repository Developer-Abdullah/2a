import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { NotificationItem } from "@/lib/types";
import SendNotificationForm from "@/components/notifications/SendNotificationForm";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

function typeVariant(t: string): BadgeVariant {
  if (t === "alert") return "default";
  if (t === "update") return "warning";
  return "muted";
}

export default async function NotificationsPage() {
  const t = await getT();
  let items: NotificationItem[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: NotificationItem[] }>("/v1/admin/notifications");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  return (
    <DashboardShell title="title.notifications">
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t("notif.send")}</CardTitle>
          </CardHeader>
          <CardContent>
            <SendNotificationForm />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("notif.sent")} ({items.length})</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {error ? (
              <p className="px-6 pb-6 text-sm text-red-600 dark:text-red-400">{error}</p>
            ) : items.length === 0 ? (
              <p className="px-6 pb-6 text-sm text-muted-foreground">{t("notif.nothing")}</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("notif.col.title")}</TableHead>
                    <TableHead>{t("notif.col.type")}</TableHead>
                    <TableHead>{t("notif.col.sent")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((n) => (
                    <TableRow key={n.id}>
                      <TableCell>
                        <div className="font-medium text-foreground">{n.title}</div>
                        <div className="max-w-xs truncate text-xs text-muted-foreground">{n.body}</div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={typeVariant(n.type)}>{n.type}</Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground">{n.sent_at.slice(0, 16)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </DashboardShell>
  );
}
