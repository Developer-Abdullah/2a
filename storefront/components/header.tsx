"use client";

import Link from "next/link";
import { ShoppingCart } from "lucide-react";
import { CurrencySwitcher } from "./currency-switcher";
import { useStore } from "./store-provider";

export function Header() {
  const { count } = useStore();
  return (
    <header className="brand-gradient sticky top-0 z-40 text-white shadow-lg">
      <div className="container-page flex h-16 items-center justify-between gap-4">
        <Link href="/" className="flex items-center gap-2.5 font-extrabold text-lg">
          <img src="/logo.jpg" alt="Double A" className="h-9 w-9 rounded-full ring-2 ring-white/40" />
          <span>Double A</span>
        </Link>

        <nav className="hidden items-center gap-6 text-sm font-semibold text-white/90 md:flex">
          <Link href="/" className="hover:text-white">الرئيسية</Link>
          <Link href="/#products" className="hover:text-white">الباقات</Link>
          <Link href="/activate" className="hover:text-white">تفعيل الكود</Link>
          <Link href="/orders" className="hover:text-white">طلباتي</Link>
        </nav>

        <div className="flex items-center gap-3">
          <div className="hidden sm:block">
            <CurrencySwitcher />
          </div>
          <Link href="/cart" className="relative rounded-xl bg-white/15 p-2 hover:bg-white/25" aria-label="السلة">
            <ShoppingCart className="h-5 w-5" />
            {count > 0 && (
              <span className="absolute -top-1 -left-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-amber-400 px-1 text-xs font-bold text-ink">
                {count}
              </span>
            )}
          </Link>
        </div>
      </div>
      <div className="container-page pb-3 sm:hidden">
        <CurrencySwitcher />
      </div>
    </header>
  );
}
