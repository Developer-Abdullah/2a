"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { CreditCard, Loader2, ShieldCheck } from "lucide-react";
import { useStore } from "@/components/store-provider";
import { Product } from "@/lib/types";
import { formatMoney, priceFor } from "@/lib/format";

export default function CheckoutPage() {
  const router = useRouter();
  const { cart, currency, clearCart, t } = useStore();
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
        body: JSON.stringify({ email, phone, currency, items: cart.map((c) => ({ slug: c.slug, qty: c.qty })) }),
      });
      const data = await res.json();
      // Manual flow: create a pending order and go to its status page (payment instructions there).
      if (res.ok && data.orderId) {
        clearCart();
        router.push(`/order/${data.orderId}`);
        return;
      }
      setError(data.error === "missing_fields" ? t("checkout.err_missing") : t("checkout.err_generic"));
    } catch {
      setError(t("checkout.err_generic"));
    } finally {
      setSubmitting(false);
    }
  };

  if (cart.length === 0) {
    return (
      <div className="container-page py-20 text-center">
        <h1 className="text-2xl font-extrabold text-foreground">{t("checkout.empty")}</h1>
        <Link href="/#products" className="btn-primary mt-6">{t("cart.empty_cta")}</Link>
      </div>
    );
  }

  return (
    <div className="container-page py-10">
      <h1 className="text-2xl font-extrabold text-foreground">{t("checkout.title")}</h1>
      <div className="mt-8 grid gap-8 lg:grid-cols-3">
        <form onSubmit={submit} className="card space-y-5 p-6 lg:col-span-2">
          <div>
            <label className="mb-1 block text-sm font-bold text-foreground/80">{t("checkout.email")} *</label>
            <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} placeholder="name@example.com" className="field" />
            <p className="mt-1 text-xs text-muted-foreground">{t("buy.note")}</p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-bold text-foreground/80">{t("checkout.phone")}</label>
            <input type="tel" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="+20 1XX XXX XXXX" className="field" />
          </div>

          <div className="flex items-start gap-2 rounded-xl bg-brand-50 p-3 text-sm text-brand-800 dark:bg-brand-900/40 dark:text-brand-100">
            <ShieldCheck className="mt-0.5 h-5 w-5 flex-none" />
            <span>{t(currency === "EGP" ? "checkout.note_egp" : "checkout.note_kwd")}</span>
          </div>

          {error && <div className="rounded-xl bg-rose-50 p-3 text-sm text-rose-600 dark:bg-rose-900/30 dark:text-rose-300">{error}</div>}

          <button type="submit" disabled={submitting} className="btn-primary w-full text-base">
            {submitting ? <Loader2 className="h-5 w-5 animate-spin" /> : <CreditCard className="h-5 w-5" />}
            {submitting ? t("checkout.creating") : `${t("checkout.confirm")} — ${formatMoney(total, currency)}`}
          </button>
        </form>

        <div className="lg:col-span-1">
          <div className="card p-6">
            <h2 className="text-lg font-extrabold text-foreground">{t("checkout.summary")}</h2>
            <ul className="mt-4 space-y-3 text-sm">
              {cart.map((item) => (
                <li key={item.slug} className="flex items-center justify-between gap-2 text-muted-foreground">
                  <span className="flex-1">{item.name} × {item.qty}</span>
                  <span className="font-bold text-foreground">{formatMoney(priceOf(item.slug) * item.qty, currency)}</span>
                </li>
              ))}
            </ul>
            <div className="mt-4 flex items-center justify-between border-t border-border pt-4">
              <span className="font-bold text-foreground/80">{t("cart.total")}</span>
              <span className="text-xl font-extrabold text-brand-600 dark:text-brand-300">{formatMoney(total, currency)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
