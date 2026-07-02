"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Currency } from "@/lib/types";

export interface CartItem {
  slug: string;
  name: string;
  qty: number;
}

interface StoreState {
  currency: Currency;
  setCurrency: (c: Currency) => void;
  cart: CartItem[];
  addToCart: (item: CartItem) => void;
  setQty: (slug: string, qty: number) => void;
  removeFromCart: (slug: string) => void;
  clearCart: () => void;
  count: number;
}

const StoreContext = createContext<StoreState | null>(null);

const CART_KEY = "store_cart";
const CURRENCY_KEY = "store_currency";

function readCart(): CartItem[] {
  if (typeof window === "undefined") return [];
  try {
    return JSON.parse(window.localStorage.getItem(CART_KEY) || "[]");
  } catch {
    return [];
  }
}

export function StoreProvider({
  initialCurrency,
  children,
}: {
  initialCurrency: Currency;
  children: React.ReactNode;
}) {
  const router = useRouter();
  const [currency, setCurrencyState] = useState<Currency>(initialCurrency);
  const [cart, setCart] = useState<CartItem[]>([]);

  useEffect(() => {
    setCart(readCart());
    const saved = window.localStorage.getItem(CURRENCY_KEY) as Currency | null;
    if (saved && saved !== initialCurrency) setCurrencyState(saved);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const persist = useCallback((next: CartItem[]) => {
    setCart(next);
    window.localStorage.setItem(CART_KEY, JSON.stringify(next));
  }, []);

  const setCurrency = useCallback(
    (c: Currency) => {
      setCurrencyState(c);
      window.localStorage.setItem(CURRENCY_KEY, c);
      // Mirror to a cookie so server components re-price on refresh.
      document.cookie = `currency=${c}; path=/; max-age=${60 * 60 * 24 * 365}`;
      router.refresh();
    },
    [router]
  );

  const addToCart = useCallback(
    (item: CartItem) => {
      const existing = cart.find((c) => c.slug === item.slug);
      const next = existing
        ? cart.map((c) => (c.slug === item.slug ? { ...c, qty: c.qty + item.qty } : c))
        : [...cart, item];
      persist(next);
    },
    [cart, persist]
  );

  const setQty = useCallback(
    (slug: string, qty: number) => {
      if (qty < 1) return;
      persist(cart.map((c) => (c.slug === slug ? { ...c, qty } : c)));
    },
    [cart, persist]
  );

  const removeFromCart = useCallback(
    (slug: string) => persist(cart.filter((c) => c.slug !== slug)),
    [cart, persist]
  );

  const clearCart = useCallback(() => persist([]), [persist]);

  const value = useMemo<StoreState>(
    () => ({
      currency,
      setCurrency,
      cart,
      addToCart,
      setQty,
      removeFromCart,
      clearCart,
      count: cart.reduce((n, c) => n + c.qty, 0),
    }),
    [currency, setCurrency, cart, addToCart, setQty, removeFromCart, clearCart]
  );

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}

export function useStore(): StoreState {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error("useStore must be used within StoreProvider");
  return ctx;
}
