"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Minus, Plus, Trash2, ShoppingCart } from "lucide-react";
import { useStore } from "@/components/store-provider";
import { Product } from "@/lib/types";
import { formatMoney, priceFor } from "@/lib/format";

export default function CartPage() {
  const { cart, currency, setQty, removeFromCart } = useStore();
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
        <ShoppingCart className="mx-auto h-14 w-14 text-slate-300" />
        <h1 className="mt-4 text-2xl font-extrabold text-ink">سلتك فارغة</h1>
        <p className="mt-2 text-slate-500">تصفّح الباقات وأضف ما يناسبك.</p>
        <Link href="/#products" className="btn-primary mt-6">تصفّح الباقات</Link>
      </div>
    );
  }

  return (
    <div className="container-page py-10">
      <h1 className="text-2xl font-extrabold text-ink">سلة المشتريات</h1>
      <div className="mt-8 grid gap-8 lg:grid-cols-3">
        <div className="space-y-4 lg:col-span-2">
          {cart.map((item) => (
            <div key={item.slug} className="card flex items-center gap-4 p-4">
              <div className="brand-gradient flex h-16 w-16 flex-none items-center justify-center rounded-xl text-white">
                <ShoppingCart className="h-6 w-6" />
              </div>
              <div className="flex-1">
                <Link href={`/${item.slug}`} className="font-bold text-ink hover:text-brand-700">{item.name}</Link>
                <div className="mt-1 text-sm text-brand-700">{loading ? "…" : formatMoney(priceOf(item.slug), currency)}</div>
              </div>
              <div className="flex items-center gap-1 rounded-xl border border-slate-200 p-1">
                <button onClick={() => setQty(item.slug, item.qty - 1)} className="rounded-lg p-1.5 hover:bg-slate-100"><Minus className="h-4 w-4" /></button>
                <span className="w-8 text-center font-bold">{item.qty}</span>
                <button onClick={() => setQty(item.slug, item.qty + 1)} className="rounded-lg p-1.5 hover:bg-slate-100"><Plus className="h-4 w-4" /></button>
              </div>
              <button onClick={() => removeFromCart(item.slug)} className="rounded-lg p-2 text-rose-500 hover:bg-rose-50" aria-label="حذف">
                <Trash2 className="h-5 w-5" />
              </button>
            </div>
          ))}
        </div>

        <div className="lg:col-span-1">
          <div className="card p-6">
            <h2 className="text-lg font-extrabold text-ink">ملخص الطلب</h2>
            <div className="mt-4 flex items-center justify-between text-slate-600">
              <span>الإجمالي</span>
              <span className="text-xl font-extrabold text-brand-700">{loading ? "…" : formatMoney(total, currency)}</span>
            </div>
            <Link href="/checkout" className="btn-primary mt-6 w-full">إتمام الشراء</Link>
            <Link href="/#products" className="btn-ghost mt-3 w-full">متابعة التسوق</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
