import Link from "next/link";
import { notFound } from "next/navigation";
import { CheckCircle2, Clock, XCircle } from "lucide-react";
import { fetchOrder } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import { paymentInfo } from "@/lib/payment";
import { getLocale, getT } from "@/lib/locale-server";
import { OrderCodes } from "@/components/order-codes";
import { OrderPoller } from "@/components/order-poller";
import { PaymentPanel } from "@/components/payment-panel";

export const dynamic = "force-dynamic";

export default async function OrderPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [locale, t] = await Promise.all([getLocale(), getT()]);
  const order = await fetchOrder(id);
  if (!order) notFound();

  const fulfilled = order.status === "fulfilled";
  const failed = order.status === "failed";
  const pending = order.status === "pending" || order.status === "paid";

  return (
    <div className="container-page max-w-3xl py-12">
      {/* While pending, poll the server so the page flips to "fulfilled" once the payment is confirmed. */}
      {pending && <OrderPoller id={order.id} />}

      <div className="card overflow-hidden">
        <div
          className={
            "flex flex-col items-center gap-3 px-6 py-10 text-center text-white " +
            (fulfilled ? "brand-gradient" : failed ? "bg-rose-500" : "bg-amber-500")
          }
        >
          {fulfilled ? <CheckCircle2 className="h-14 w-14" /> : failed ? <XCircle className="h-14 w-14" /> : <Clock className="h-14 w-14" />}
          <h1 className="text-2xl font-extrabold">
            {fulfilled ? t("order.fulfilled_title") : failed ? t("order.failed_title") : t("order.pending_title")}
          </h1>
          <p className="text-white/90">
            {fulfilled ? t("order.fulfilled_sub") : failed ? t("order.failed_sub") : t("order.pending_sub")}
          </p>
        </div>

        <div className="space-y-6 p-6">
          <div className="flex items-center justify-between gap-2 text-sm text-muted-foreground">
            <span>{t("order.number")}</span>
            <span className="break-all font-mono text-foreground/80">{order.id}</span>
          </div>

          <div className="rounded-2xl bg-muted p-4">
            <h2 className="font-extrabold text-foreground">{t("order.details")}</h2>
            <ul className="mt-3 space-y-2 text-sm">
              {order.items.map((it) => (
                <li key={it.id} className="flex items-center justify-between gap-2 text-muted-foreground">
                  <span>{it.product_name} × {it.qty}</span>
                  <span className="font-bold text-foreground">{formatMoney(it.unit_amount * it.qty, it.currency)}</span>
                </li>
              ))}
            </ul>
            <div className="mt-3 flex items-center justify-between border-t border-border pt-3">
              <span className="font-bold text-foreground/80">{t("order.total")}</span>
              <span className="text-lg font-extrabold text-brand-600 dark:text-brand-300">{formatMoney(order.total, order.currency)}</span>
            </div>
          </div>

          {fulfilled && (
            <>
              <OrderCodes codes={order.codes} />
              <Link href="/activate" className="btn-primary w-full">{t("order.activate_now")}</Link>
            </>
          )}

          {pending && (() => {
            const pay = paymentInfo(order.currency, locale);
            return (
              <PaymentPanel
                orderId={order.id}
                total={formatMoney(order.total, order.currency)}
                methods={pay.methods}
                whatsapp={pay.whatsapp}
                hasProof={order.has_payment_proof}
              />
            );
          })()}

          <div className="flex justify-center">
            <Link href="/" className="btn-ghost">{t("order.back_home")}</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
