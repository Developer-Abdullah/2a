import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { NotificationItem } from "@/lib/types";
import SendNotificationForm from "@/components/notifications/SendNotificationForm";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

function typeVariant(t: string): BadgeVariant {
  if (t === "alert") return "default";
  if (t === "update") return "warning";
  return "muted";
}

export default async function NotificationsPage() {
  let items: NotificationItem[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: NotificationItem[] }>("/v1/admin/notifications");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load notifications";
  }

  return (
    <DashboardShell title="title.notifications">
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Send notification</CardTitle>
          </CardHeader>
          <CardContent>
            <SendNotificationForm />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Sent ({items.length})</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {error ? (
              <p className="px-6 pb-6 text-sm text-red-600">{error}</p>
            ) : items.length === 0 ? (
              <p className="px-6 pb-6 text-sm text-slate-500">Nothing sent yet.</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Sent</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((n) => (
                    <TableRow key={n.id}>
                      <TableCell>
                        <div className="font-medium text-slate-900">{n.title}</div>
                        <div className="max-w-xs truncate text-xs text-slate-500">{n.body}</div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={typeVariant(n.type)}>{n.type}</Badge>
                      </TableCell>
                      <TableCell className="text-slate-500">{n.sent_at.slice(0, 16)}</TableCell>
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
