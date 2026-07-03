import Link from "next/link";
import { ChevronRight } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { OrderDetail } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import ConfirmOrderButton from "@/components/orders/ConfirmOrderButton";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

const STATUS_VARIANT: Record<string, BadgeVariant> = {
  pending: "warning",
  paid: "default",
  failed: "danger",
  fulfilled: "success",
};

export default async function OrderDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const t = await getT();

  let order: OrderDetail | null = null;
  let error = "";
  try {
    const data = await adminGet<{ order: OrderDetail }>(`/v1/admin/orders/${id}`);
    order = data.order;
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  return (
    <DashboardShell title="title.order_details">
      <Link href="/orders" className="mb-6 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ChevronRight className="h-4 w-4 rtl:rotate-180" /> {t("order.back")}
      </Link>

      {error || !order ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error || t("order.notfound")}</p>
      ) : (
        <div className="grid gap-6 lg:grid-cols-3">
          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>#{order.id.slice(0, 8)}</span>
                <Badge variant={STATUS_VARIANT[order.status] ?? "muted"}>{t(`orders.status.${order.status}`)}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm text-muted-foreground">
              <div className="grid grid-cols-2 gap-3">
                <Info label={t("order.email")} value={order.email} />
                <Info label={t("order.phone")} value={order.phone || "—"} />
                <Info label={t("order.currency")} value={order.currency} />
                <Info label={t("order.gateway")} value={order.provider || "—"} />
                <Info label={t("order.date")} value={new Date(order.created_at).toLocaleString()} />
                <Info label={t("order.total")} value={`${order.total} ${order.currency}`} />
              </div>

              <div>
                <h3 className="mb-2 font-semibold text-foreground">{t("order.items")}</h3>
                <ul className="space-y-1">
                  {order.items.map((it) => (
                    <li key={it.id} className="flex items-center justify-between border-b border-border py-1">
                      <span>{it.product_name} × {it.qty}</span>
                      <span className="font-medium text-foreground/90">{it.unit_amount * it.qty} {it.currency}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </CardContent>
          </Card>

          <Card className="h-fit">
            <CardHeader>
              <CardTitle>{t("order.codes_title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {order.status !== "fulfilled" && (
                <div className="space-y-3 rounded-lg bg-amber-50 p-3 dark:bg-amber-900/20">
                  {order.has_payment_proof ? (
                    <a href={`/api/order-proof/${order.id}`} target="_blank" rel="noopener noreferrer" className="block">
                      <div className="mb-1 text-xs font-medium text-emerald-700 dark:text-emerald-400">{t("order.proof_uploaded")}</div>
                      <img src={`/api/order-proof/${order.id}`} alt="proof" className="max-h-56 w-full rounded-lg object-contain ring-1 ring-border" />
                    </a>
                  ) : (
                    <p className="text-sm text-amber-800 dark:text-amber-300">{t("order.no_proof")}</p>
                  )}
                  <p className="text-sm text-amber-800 dark:text-amber-300">{t("order.confirm_hint")}</p>
                  <ConfirmOrderButton id={order.id} />
                </div>
              )}

              {order.codes.length === 0 ? (
                order.status === "fulfilled" ? <p className="text-sm text-muted-foreground">{t("order.no_codes")}</p> : null
              ) : (
                <ul className="space-y-2">
                  {order.codes.map((code) => (
                    <li key={code} className="rounded-lg bg-muted px-3 py-2 font-mono text-sm font-bold tracking-wider text-foreground">
                      {code}
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="font-medium text-foreground/90">{value}</div>
    </div>
  );
}
