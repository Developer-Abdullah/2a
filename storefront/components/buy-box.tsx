"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Minus, Plus, ShoppingCart, Zap } from "lucide-react";
import { Product } from "@/lib/types";
import { useStore } from "./store-provider";
import { Price } from "./price";

export function BuyBox({ product }: { product: Product }) {
  const router = useRouter();
  const { addToCart, t } = useStore();
  const [qty, setQty] = useState(1);
  const [added, setAdded] = useState(false);

  const dec = () => setQty((q) => Math.max(1, q - 1));
  const inc = () => setQty((q) => Math.min(10, q + 1));

  const add = () => {
    addToCart({ slug: product.slug, name: product.name, qty });
    setAdded(true);
    setTimeout(() => setAdded(false), 1800);
  };

  const buyNow = () => {
    addToCart({ slug: product.slug, name: product.name, qty });
    router.push("/checkout");
  };

  return (
    <div className="card p-5">
      <Price product={product} size="lg" />

      <div className="mt-5 flex items-center justify-between">
        <span className="text-sm font-bold text-muted-foreground">{t("buy.qty")}</span>
        <div className="flex items-center gap-1 rounded-xl border border-border p-1">
          <button onClick={dec} className="rounded-lg p-2 text-muted-foreground hover:bg-muted" aria-label="-"><Minus className="h-4 w-4" /></button>
          <span className="w-10 text-center font-extrabold text-foreground">{qty}</span>
          <button onClick={inc} className="rounded-lg p-2 text-muted-foreground hover:bg-muted" aria-label="+"><Plus className="h-4 w-4" /></button>
        </div>
      </div>

      <div className="mt-5 grid gap-3">
        <button onClick={buyNow} className="btn-primary w-full text-base">
          <Zap className="h-5 w-5" /> {t("buy.buynow")}
        </button>
        <button onClick={add} className="btn-ghost w-full text-base">
          <ShoppingCart className="h-5 w-5" /> {added ? t("buy.added") : t("buy.add")}
        </button>
      </div>

      <p className="mt-4 text-center text-xs text-muted-foreground">{t("buy.note")}</p>
    </div>
  );
}
