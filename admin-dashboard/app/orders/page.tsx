import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { OrderSummary } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

const STATUS_VARIANT: Record<string, BadgeVariant> = {
  pending: "warning",
  paid: "default",
  failed: "danger",
  fulfilled: "success",
};

export default async function OrdersPage() {
  const t = await getT();
  let orders: OrderSummary[] = [];
  let error = "";
  try {
    const data = await adminGet<{ items: OrderSummary[] }>("/v1/admin/orders");
    orders = data.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  const statusBadge = (status: string) => (
    <Badge variant={STATUS_VARIANT[status] ?? "muted"}>{t(`orders.status.${status}`)}</Badge>
  );

  return (
    <DashboardShell title="orders.title">
      <p className="mb-6 text-sm text-muted-foreground">{orders.length} {t("orders.count")}</p>

      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : orders.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("orders.count")}: 0</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("orders.col.email")}</TableHead>
                  <TableHead>{t("orders.col.total")}</TableHead>
                  <TableHead>{t("orders.col.codes")}</TableHead>
                  <TableHead>{t("orders.col.status")}</TableHead>
                  <TableHead className="text-end"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {orders.map((o) => (
                  <TableRow key={o.id}>
                    <TableCell className="font-medium text-foreground">{o.email}</TableCell>
                    <TableCell className="text-foreground/80">{o.total} {o.currency}</TableCell>
                    <TableCell className="text-foreground/80">{o.code_count}</TableCell>
                    <TableCell>{statusBadge(o.status)}</TableCell>
                    <TableCell className="text-end">
                      <Link href={`/orders/${o.id}`} className="text-sm font-medium text-brand-700 hover:underline dark:text-brand-300">
                        {t("common.edit")}
                      </Link>
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
