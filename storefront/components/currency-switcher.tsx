"use client";

import { CURRENCIES, CURRENCY_LABELS } from "@/lib/types";
import { useStore } from "./store-provider";
import { clsx } from "clsx";

export function CurrencySwitcher() {
  const { currency, setCurrency } = useStore();
  return (
    <div className="inline-flex rounded-xl bg-white/15 p-1 text-sm font-bold text-white">
      {CURRENCIES.map((c) => (
        <button
          key={c}
          onClick={() => setCurrency(c)}
          className={clsx(
            "rounded-lg px-3 py-1 transition",
            currency === c ? "bg-white text-brand-700" : "text-white/90 hover:text-white"
          )}
        >
          {c} {CURRENCY_LABELS[c]}
        </button>
      ))}
    </div>
  );
}
