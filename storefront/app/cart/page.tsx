"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Minus, Plus, Trash2, ShoppingCart } from "lucide-react";
import { useStore } from "@/components/store-provider";
import { Product } from "@/lib/types";
import { formatMoney, priceFor } from "@/lib/format";

export default function CartPage() {
  const { cart, currency, setQty, removeFromCart, t } = useStore();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetch(`/api/products?currency=${currency}`)
      .then((r) => r.json())
      .then((d) => setProducts(d.products ?? []))
      .catch(() => setProducts([]))
      .finally(() => setLoading(false));
  }, [currency]);

  const priceOf = (slug: string) => {
    const p = products.find((x) => x.slug === slug);
    return p ? priceFor(p.prices, currency)?.amount ?? 0 : 0;
  };
  const total = cart.reduce((sum, item) => sum + priceOf(item.slug) * item.qty, 0);

  if (cart.length === 0) {
    return (
      <div className="container-page py-20 text-center">
        <ShoppingCart className="mx-auto h-14 w-14 text-muted-foreground/40" />
        <h1 className="mt-4 text-2xl font-extrabold text-foreground">{t("cart.empty")}</h1>
        <p className="mt-2 text-muted-foreground">{t("home.products_sub")}</p>
        <Link href="/#products" className="btn-primary mt-6">{t("cart.empty_cta")}</Link>
      </div>
    );
  }

  return (
    <div className="container-page py-10">
      <h1 className="text-2xl font-extrabold text-foreground">{t("cart.title")}</h1>
      <div className="mt-8 grid gap-8 lg:grid-cols-3">
        <div className="space-y-4 lg:col-span-2">
          {cart.map((item) => (
            <div key={item.slug} className="card flex items-center gap-3 p-4 sm:gap-4">
              <div className="brand-gradient flex h-14 w-14 flex-none items-center justify-center rounded-xl text-white sm:h-16 sm:w-16">
                <ShoppingCart className="h-6 w-6" />
              </div>
              <div className="min-w-0 flex-1">
                <Link href={`/${item.slug}`} className="block truncate font-bold text-foreground hover:text-brand-600">{item.name}</Link>
                <div className="mt-1 text-sm text-brand-600 dark:text-brand-300">{loading ? "…" : formatMoney(priceOf(item.slug), currency)}</div>
              </div>
              <div className="flex items-center gap-1 rounded-xl border border-border p-1">
                <button onClick={() => setQty(item.slug, item.qty - 1)} className="rounded-lg p-1.5 text-muted-foreground hover:bg-muted"><Minus className="h-4 w-4" /></button>
                <span className="w-8 text-center font-bold text-foreground">{item.qty}</span>
                <button onClick={() => setQty(item.slug, item.qty + 1)} className="rounded-lg p-1.5 text-muted-foreground hover:bg-muted"><Plus className="h-4 w-4" /></button>
              </div>
              <button onClick={() => removeFromCart(item.slug)} className="rounded-lg p-2 text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-900/30" aria-label={t("cart.remove")}>
                <Trash2 className="h-5 w-5" />
              </button>
            </div>
          ))}
        </div>

        <div className="lg:col-span-1">
          <div className="card p-6">
            <h2 className="text-lg font-extrabold text-foreground">{t("checkout.summary")}</h2>
            <div className="mt-4 flex items-center justify-between text-muted-foreground">
              <span>{t("cart.total")}</span>
              <span className="text-xl font-extrabold text-brand-600 dark:text-brand-300">{loading ? "…" : formatMoney(total, currency)}</span>
            </div>
            <Link href="/checkout" className="btn-primary mt-6 w-full">{t("cart.checkout")}</Link>
            <Link href="/#products" className="btn-ghost mt-3 w-full">{t("cart.continue")}</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
