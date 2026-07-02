"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { CreditCard, Loader2, ShieldCheck } from "lucide-react";
import { useStore } from "@/components/store-provider";
import { Product } from "@/lib/types";
import { CURRENCY_LABELS, Currency } from "@/lib/types";
import { formatMoney, priceFor } from "@/lib/format";

export default function CheckoutPage() {
  const router = useRouter();
  const { cart, currency, clearCart } = useStore();
  const [products, setProducts] = useState<Product[]>([]);
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch(`/api/products?currency=${currency}`)
      .then((r) => r.json())
      .then((d) => setProducts(d.products ?? []))
      .catch(() => setProducts([]));
  }, [currency]);

  const priceOf = (slug: string) => {
    const p = products.find((x) => x.slug === slug);
    return p ? priceFor(p.prices, currency)?.amount ?? 0 : 0;
  };
  const total = cart.reduce((sum, item) => sum + priceOf(item.slug) * item.qty, 0);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const res = await fetch("/api/checkout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email,
          phone,
          currency,
          items: cart.map((c) => ({ slug: c.slug, qty: c.qty })),
        }),
      });
      const data = await res.json();
      // Manual flow: the order is created as pending; send the customer to its status page where the
      // payment instructions are shown. An admin confirms the payment to release the codes.
      if (res.ok && data.orderId) {
        clearCart();
        router.push(`/order/${data.orderId}`);
        return;
      }
      setError(messageFor(data.error));
    } catch {
      setError("تعذّر الاتصال بالخادم. حاول مرة أخرى.");
    } finally {
      setSubmitting(false);
    }
  };

  if (cart.length === 0) {
    return (
      <div className="container-page py-20 text-center">
        <h1 className="text-2xl font-extrabold text-ink">لا يوجد ما تدفع مقابله</h1>
        <Link href="/#products" className="btn-primary mt-6">تصفّح الباقات</Link>
      </div>
    );
  }

  return (
    <div className="container-page py-10">
      <h1 className="text-2xl font-extrabold text-ink">إتمام الشراء</h1>
      <div className="mt-8 grid gap-8 lg:grid-cols-3">
        <form onSubmit={submit} className="card space-y-5 p-6 lg:col-span-2">
          <div>
            <label className="mb-1 block text-sm font-bold text-slate-700">البريد الإلكتروني *</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="name@example.com"
              className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            />
            <p className="mt-1 text-xs text-slate-400">سيصلك الكود على هذا البريد وعلى صفحة الطلب.</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-bold text-slate-700">رقم الجوال (اختياري)</label>
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="+20 1XX XXX XXXX"
              className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
            />
          </div>

          <div className="flex items-start gap-2 rounded-xl bg-brand-50 p-3 text-sm text-brand-800">
            <ShieldCheck className="mt-0.5 h-5 w-5 flex-none" />
            <span>
              بعد تأكيد الطلب ستظهر لك تعليمات الدفع
              {currency === "EGP" ? " (مصر)" : " (الكويت)"}. حوّل المبلغ ثم تواصل معنا لتأكيد طلبك، ويصلك الكود فور المراجعة.
            </span>
          </div>

          {error && <div className="rounded-xl bg-rose-50 p-3 text-sm text-rose-600">{error}</div>}

          <button type="submit" disabled={submitting} className="btn-primary w-full text-base">
            {submitting ? <Loader2 className="h-5 w-5 animate-spin" /> : <CreditCard className="h-5 w-5" />}
            {submitting ? "جارٍ إنشاء الطلب…" : `تأكيد الطلب — ${formatMoney(total, currency)}`}
          </button>
        </form>

        <div className="lg:col-span-1">
          <div className="card p-6">
            <h2 className="text-lg font-extrabold text-ink">ملخص الطلب</h2>
            <ul className="mt-4 space-y-3 text-sm">
              {cart.map((item) => (
                <li key={item.slug} className="flex items-center justify-between gap-2 text-slate-600">
                  <span className="flex-1">{item.name} × {item.qty}</span>
                  <span className="font-bold text-ink">{formatMoney(priceOf(item.slug) * item.qty, currency)}</span>
                </li>
              ))}
            </ul>
            <div className="mt-4 flex items-center justify-between border-t border-slate-100 pt-4">
              <span className="font-bold text-slate-700">الإجمالي ({CURRENCY_LABELS[currency as Currency]})</span>
              <span className="text-xl font-extrabold text-brand-700">{formatMoney(total, currency)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function messageFor(code?: string): string {
  switch (code) {
    case "unsupported_currency":
      return "العملة غير مدعومة حاليًا.";
    case "empty_cart":
      return "السلة فارغة.";
    case "checkout_failed":
      return "تعذّر بدء الدفع. بوابة الدفع غير مهيّأة حاليًا — تم حفظ طلبك.";
    default:
      return "حدث خطأ أثناء إنشاء الطلب. حاول مرة أخرى.";
  }
}
