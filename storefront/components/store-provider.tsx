"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Currency } from "@/lib/types";
import { Locale, LOCALE_COOKIE, dir, translate } from "@/lib/i18n";

export type Theme = "light" | "dark";

export interface CartItem {
  slug: string;
  name: string;
  qty: number;
}

interface StoreState {
  currency: Currency;
  setCurrency: (c: Currency) => void;
  locale: Locale;
  setLocale: (l: Locale) => void;
  toggleLocale: () => void;
  t: (key: string) => string;
  theme: Theme;
  toggleTheme: () => void;
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
const THEME_KEY = "store_theme";

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
  initialLocale,
  children,
}: {
  initialCurrency: Currency;
  initialLocale: Locale;
  children: React.ReactNode;
}) {
  const router = useRouter();
  const [currency, setCurrencyState] = useState<Currency>(initialCurrency);
  const [locale, setLocaleState] = useState<Locale>(initialLocale);
  const [theme, setThemeState] = useState<Theme>("light");
  const [cart, setCart] = useState<CartItem[]>([]);

  useEffect(() => {
    setCart(readCart());
    const savedCur = window.localStorage.getItem(CURRENCY_KEY) as Currency | null;
    if (savedCur && savedCur !== initialCurrency) setCurrencyState(savedCur);
    // Theme is client-only (an inline script in <head> applies the class before paint).
    const stored = (window.localStorage.getItem(THEME_KEY) as Theme) || null;
    setThemeState(stored || (document.documentElement.classList.contains("dark") ? "dark" : "light"));
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

  const setLocale = useCallback(
    (l: Locale) => {
      setLocaleState(l);
      document.documentElement.lang = l;
      document.documentElement.dir = dir(l);
      document.cookie = `${LOCALE_COOKIE}=${l}; path=/; max-age=${60 * 60 * 24 * 365}; samesite=lax`;
      // Re-render server components (which translate from the cookie) with the new language.
      router.refresh();
    },
    [router]
  );

  const toggleLocale = useCallback(() => setLocale(locale === "ar" ? "en" : "ar"), [locale, setLocale]);
  const t = useCallback((key: string) => translate(locale, key), [locale]);

  const toggleTheme = useCallback(() => {
    const next: Theme = theme === "dark" ? "light" : "dark";
    setThemeState(next);
    window.localStorage.setItem(THEME_KEY, next);
    document.documentElement.classList.toggle("dark", next === "dark");
  }, [theme]);

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
      locale,
      setLocale,
      toggleLocale,
      t,
      theme,
      toggleTheme,
      cart,
      addToCart,
      setQty,
      removeFromCart,
      clearCart,
      count: cart.reduce((n, c) => n + c.qty, 0),
    }),
    [currency, setCurrency, locale, setLocale, toggleLocale, t, theme, toggleTheme, cart, addToCart, setQty, removeFromCart, clearCart]
  );

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}

export function useStore(): StoreState {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error("useStore must be used within StoreProvider");
  return ctx;
}
