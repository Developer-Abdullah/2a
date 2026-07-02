import { CURRENCY_LABELS, Currency, ProductPrice } from "./types";

// KWD is a 3-decimal currency; EGP shows 2. Format accordingly and append the Arabic currency label.
export function formatMoney(amount: number, currency: string): string {
  const decimals = currency === "KWD" ? 3 : 2;
  const label = CURRENCY_LABELS[currency as Currency] ?? currency;
  const value = amount.toLocaleString("ar-EG", {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
  return `${value} ${label}`;
}

export function priceFor(prices: ProductPrice[], currency: string): ProductPrice | undefined {
  return prices.find((p) => p.currency === currency) ?? prices[0];
}

export function discountPercent(price: ProductPrice): number | null {
  if (!price.compare_at || price.compare_at <= price.amount) return null;
  return Math.round((1 - price.amount / price.compare_at) * 100);
}
