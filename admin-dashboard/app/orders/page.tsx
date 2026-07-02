import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { OrderSummary } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

const STATUS: Record<string, { label: string; variant: BadgeVariant }> = {
  pending: { label: "بانتظار الدفع", variant: "warning" },
  paid: { label: "مدفوع", variant: "default" },
  failed: { label: "فشل", variant: "danger" },
  fulfilled: { label: "مكتمل", variant: "success" },
};

function statusBadge(status: string) {
  const s = STATUS[status] ?? { label: status, variant: "muted" as BadgeVariant };
  return <Badge variant={s.variant}>{s.label}</Badge>;
}

export default async function OrdersPage() {
  let orders: OrderSummary[] = [];
  let error = "";
  try {
    const data = await adminGet<{ items: OrderSummary[] }>("/v1/admin/orders");
    orders = data.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "تعذّر تحميل الطلبات";
  }

  return (
    <DashboardShell title="الطلبات">
      <p className="mb-6 text-sm text-slate-500">{orders.length} طلب</p>

      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : orders.length === 0 ? (
        <p className="text-sm text-slate-500">لا توجد طلبات بعد.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>البريد</TableHead>
                  <TableHead>الإجمالي</TableHead>
                  <TableHead>الأكواد</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead>التاريخ</TableHead>
                  <TableHead className="text-right">إجراء</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {orders.map((o) => (
                  <TableRow key={o.id}>
                    <TableCell className="font-medium text-slate-900">{o.email}</TableCell>
                    <TableCell className="text-slate-600">{o.total} {o.currency}</TableCell>
                    <TableCell className="text-slate-600">{o.code_count}</TableCell>
                    <TableCell>{statusBadge(o.status)}</TableCell>
                    <TableCell className="text-slate-500">{new Date(o.created_at).toLocaleString("ar-EG")}</TableCell>
                    <TableCell className="text-right">
                      <Link href={`/orders/${o.id}`} className="text-sm font-medium text-blue-600 hover:underline">
                        عرض
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
