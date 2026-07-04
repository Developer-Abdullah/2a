"use client";

import { Product } from "@/lib/types";
import { discountPercent, formatMoney, priceFor } from "@/lib/format";
import { useStore } from "./store-provider";

// Client price so it reacts to the currency switcher without a server refetch (product carries all
// currency prices). `staticCurrency` overrides the context, used where currency is fixed (e.g. cart).
export function Price({
  product,
  staticCurrency,
  size = "md",
}: {
  product: Pick<Product, "prices">;
  staticCurrency?: string;
  size?: "md" | "lg";
}) {
  const { currency, locale } = useStore();
  const cur = staticCurrency ?? currency;
  const price = priceFor(product.prices, cur);
  if (!price) return null;
  const off = discountPercent(price);

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className={(size === "lg" ? "text-3xl" : "text-xl") + " font-extrabold text-brand-600 dark:text-brand-300"}>
        {formatMoney(price.amount, price.currency)}
      </span>
      {price.compare_at && off ? (
        <>
          <span className="text-sm text-muted-foreground line-through">
            {formatMoney(price.compare_at, price.currency)}
          </span>
          <span className="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-bold text-rose-600 dark:bg-rose-900/40 dark:text-rose-300">
            {locale === "ar" ? `خصم ${off}%` : `${off}% off`}
          </span>
        </>
      ) : null}
    </div>
  );
}
