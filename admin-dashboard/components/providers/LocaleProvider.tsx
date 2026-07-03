"use client";

import { createContext, useCallback, useContext, useState } from "react";
import { useRouter } from "next/navigation";
import { DEFAULT_LOCALE, LOCALE_COOKIE, Locale, dir, translate } from "@/lib/i18n";

const LocaleContext = createContext<{
  locale: Locale;
  toggle: () => void;
  setLocale: (l: Locale) => void;
  t: (key: string) => string;
}>({
  locale: DEFAULT_LOCALE,
  toggle: () => {},
  setLocale: () => {},
  t: (k) => k,
});

// Holds the active locale, flips <html dir/lang>, and persists the choice in a cookie so server
// components can read it on the next request. `initial` comes from the server (cookie) to avoid a
// hydration mismatch.
export function LocaleProvider({ initial, children }: { initial: Locale; children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(initial);
  const router = useRouter();

  const setLocale = useCallback(
    (l: Locale) => {
      setLocaleState(l);
      if (typeof document !== "undefined") {
        document.documentElement.lang = l;
        document.documentElement.dir = dir(l);
        // 1-year cookie so SSR picks up the language on subsequent navigations.
        document.cookie = `${LOCALE_COOKIE}=${l}; path=/; max-age=31536000; samesite=lax`;
      }
      // Re-render server components (which translate from the cookie) with the new language.
      router.refresh();
    },
    [router]
  );

  const toggle = useCallback(() => setLocale(locale === "ar" ? "en" : "ar"), [locale, setLocale]);
  const t = useCallback((key: string) => translate(locale, key), [locale]);

  return <LocaleContext.Provider value={{ locale, toggle, setLocale, t }}>{children}</LocaleContext.Provider>;
}

export const useI18n = () => useContext(LocaleContext);
