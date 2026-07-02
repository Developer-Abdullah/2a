import Link from "next/link";
import { notFound } from "next/navigation";
import { CheckCircle2, Clock, XCircle, Wallet, MessageCircle } from "lucide-react";
import { fetchOrder } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import { paymentInfo } from "@/lib/payment";
import { OrderCodes } from "@/components/order-codes";
import { OrderPoller } from "@/components/order-poller";

export const dynamic = "force-dynamic";

export default async function OrderPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const order = await fetchOrder(id);
  if (!order) notFound();

  const fulfilled = order.status === "fulfilled";
  const failed = order.status === "failed";
  const pending = order.status === "pending" || order.status === "paid";

  return (
    <div className="container-page max-w-3xl py-12">
      {/* While pending, poll the server so the page flips to "fulfilled" once the webhook lands. */}
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
            {fulfilled ? "تم استلام الطلب 🎉" : failed ? "الطلب ملغى" : "طلبك قيد المراجعة"}
          </h1>
          <p className="text-white/90">
            {fulfilled
              ? "شكرًا لك! كود التفعيل جاهز بالأسفل."
              : failed
              ? "تم إلغاء هذا الطلب. يمكنك إنشاء طلب جديد."
              : "أكمل الدفع بالتعليمات بالأسفل، وسيصلك الكود فور تأكيد الدفع."}
          </p>
        </div>

        <div className="space-y-6 p-6">
          <div className="flex items-center justify-between text-sm text-slate-500">
            <span>رقم الطلب</span>
            <span className="font-mono text-slate-700">{order.id}</span>
          </div>

          <div className="rounded-2xl bg-slate-50 p-4">
            <h2 className="font-extrabold text-ink">تفاصيل الطلب</h2>
            <ul className="mt-3 space-y-2 text-sm">
              {order.items.map((it) => (
                <li key={it.id} className="flex items-center justify-between text-slate-600">
                  <span>{it.product_name} × {it.qty}</span>
                  <span className="font-bold text-ink">{formatMoney(it.unit_amount * it.qty, it.currency)}</span>
                </li>
              ))}
            </ul>
            <div className="mt-3 flex items-center justify-between border-t border-slate-200 pt-3">
              <span className="font-bold text-slate-700">الإجمالي</span>
              <span className="text-lg font-extrabold text-brand-700">{formatMoney(order.total, order.currency)}</span>
            </div>
          </div>

          {fulfilled && <OrderCodes codes={order.codes} />}

          {pending && <PaymentPanel currency={order.currency} orderId={order.id} total={formatMoney(order.total, order.currency)} />}

          <div className="flex justify-center">
            <Link href="/" className="btn-ghost">العودة للرئيسية</Link>
          </div>
        </div>
      </div>
    </div>
  );
}

// PaymentPanel shows manual payment instructions for a pending order: how to pay for the selected
// currency, the order id to use as a transfer reference, and a WhatsApp button to send proof.
function PaymentPanel({ currency, orderId, total }: { currency: string; orderId: string; total: string }) {
  const { instructions, whatsapp } = paymentInfo(currency);
  const waText = encodeURIComponent(`مرحبًا، أودّ تأكيد دفع الطلب رقم ${orderId} بقيمة ${total}.`);
  const waLink = whatsapp ? `https://wa.me/${whatsapp}?text=${waText}` : "";

  return (
    <div className="rounded-2xl border-2 border-dashed border-amber-300 bg-amber-50 p-5">
      <h2 className="flex items-center gap-2 font-extrabold text-amber-900">
        <Wallet className="h-5 w-5" /> تعليمات الدفع
      </h2>
      <p className="mt-3 text-sm leading-7 text-amber-900">{instructions}</p>
      <div className="mt-3 rounded-xl bg-white p-3 text-sm ring-1 ring-amber-100">
        <div className="flex items-center justify-between">
          <span className="text-slate-500">المبلغ</span>
          <span className="font-extrabold text-ink">{total}</span>
        </div>
        <div className="mt-1 flex items-center justify-between">
          <span className="text-slate-500">رقم الطلب (المرجع)</span>
          <span className="font-mono text-slate-700">{orderId}</span>
        </div>
      </div>
      {waLink && (
        <a href={waLink} target="_blank" rel="noopener noreferrer" className="btn-primary mt-4 w-full">
          <MessageCircle className="h-5 w-5" /> أرسل إثبات الدفع عبر واتساب
        </a>
      )}
      <p className="mt-3 text-center text-xs text-amber-700/80">
        بعد تأكيد دفعك سيظهر كود التفعيل هنا تلقائيًا. أبقِ الصفحة مفتوحة.
      </p>
    </div>
  );
}
