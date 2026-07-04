"use client";

import Link from "next/link";
import { useState } from "react";
import { ShoppingCart, Menu, X, Moon, Sun, Languages } from "lucide-react";
import { CurrencySwitcher } from "./currency-switcher";
import { useStore } from "./store-provider";

export function Header() {
  const { count, t, locale, toggleLocale, theme, toggleTheme } = useStore();
  const [open, setOpen] = useState(false);

  const nav = [
    { href: "/", label: t("nav.home") },
    { href: "/#products", label: t("nav.packages") },
    { href: "/activate", label: t("nav.activate") },
    { href: "/orders", label: t("nav.orders") },
  ];

  const IconBtn = ({ onClick, label, children }: { onClick: () => void; label: string; children: React.ReactNode }) => (
    <button onClick={onClick} aria-label={label} title={label} className="rounded-xl bg-white/15 p-2 text-white hover:bg-white/25">
      {children}
    </button>
  );

  return (
    <header className="brand-gradient sticky top-0 z-40 text-white shadow-lg">
      <div className="container-page flex h-16 items-center justify-between gap-3">
        <Link href="/" className="flex items-center gap-2.5 text-lg font-extrabold" onClick={() => setOpen(false)}>
          <img src="/logo.jpg" alt="Double A" className="h-9 w-9 rounded-full ring-2 ring-white/40" />
          <span>Double A</span>
        </Link>

        <nav className="hidden items-center gap-6 text-sm font-semibold text-white/90 md:flex">
          {nav.map((n) => (
            <Link key={n.href} href={n.href} className="hover:text-white">{n.label}</Link>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          <div className="hidden sm:block">
            <CurrencySwitcher />
          </div>
          <IconBtn onClick={toggleLocale} label={locale === "ar" ? "English" : "عربي"}>
            <span className="flex items-center gap-1 text-sm font-bold">
              <Languages className="h-4 w-4" /> {locale === "ar" ? "EN" : "ع"}
            </span>
          </IconBtn>
          <IconBtn onClick={toggleTheme} label={theme === "dark" ? "Light" : "Dark"}>
            {theme === "dark" ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
          </IconBtn>
          <Link href="/cart" className="relative rounded-xl bg-white/15 p-2 hover:bg-white/25" aria-label={t("nav.cart")}>
            <ShoppingCart className="h-5 w-5" />
            {count > 0 && (
              <span className="absolute -top-1 -start-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-amber-400 px-1 text-xs font-bold text-ink">
                {count}
              </span>
            )}
          </Link>
          {/* Mobile menu toggle */}
          <button onClick={() => setOpen((v) => !v)} aria-label={t("nav.menu")} className="rounded-xl bg-white/15 p-2 hover:bg-white/25 md:hidden">
            {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>
      </div>

      {/* Mobile panel */}
      {open && (
        <div className="border-t border-white/15 md:hidden">
          <nav className="container-page flex flex-col gap-1 py-3 text-sm font-semibold">
            {nav.map((n) => (
              <Link key={n.href} href={n.href} onClick={() => setOpen(false)} className="rounded-lg px-3 py-2 hover:bg-white/10">
                {n.label}
              </Link>
            ))}
            <div className="px-1 pt-2">
              <CurrencySwitcher />
            </div>
          </nav>
        </div>
      )}
    </header>
  );
}
