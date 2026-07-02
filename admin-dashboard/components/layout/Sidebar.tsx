"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Package,
  ShoppingCart,
  AppWindow,
  Star,
  Bell,
  KeyRound,
  Users,
  Smartphone,
  BarChart3,
  ShieldCheck,
  Settings,
} from "lucide-react";
import { cn } from "@/lib/utils";

// Single-store admin nav. The storefront catalog (products) and orders are the primary surfaces; the
// older per-merchant "owner console" has been retired now that this runs as one store.
const storeNav = [
  { href: "/dashboard", label: "لوحة التحكم", icon: LayoutDashboard },
  { href: "/products", label: "المنتجات", icon: Package },
  { href: "/orders", label: "الطلبات", icon: ShoppingCart },
  { href: "/apps", label: "التطبيقات", icon: AppWindow },
  { href: "/activation", label: "الأكواد", icon: KeyRound },
  { href: "/ratings", label: "التقييمات", icon: Star },
  { href: "/notifications", label: "الإشعارات", icon: Bell },
  { href: "/users", label: "العملاء", icon: Users },
  { href: "/devices", label: "الأجهزة", icon: Smartphone },
  { href: "/analytics", label: "الإحصائيات", icon: BarChart3 },
  { href: "/enrollment", label: "التسجيل", icon: ShieldCheck },
  { href: "/settings", label: "الإعدادات", icon: Settings },
];

// isOwner is accepted for backward compatibility but no longer switches the nav — there is one store.
export default function Sidebar({ isOwner: _isOwner = false }: { isOwner?: boolean }) {
  const pathname = usePathname();
  const nav = storeNav;

  return (
    <aside dir="rtl" className="w-64 shrink-0 border-l border-slate-200 bg-white">
      <div className="flex h-16 items-center gap-2 border-b border-slate-200 px-6">
        <img src="/logo.jpg" alt="Double A" className="h-8 w-8 rounded-md" />
        <div className="leading-tight">
          <div className="text-sm font-semibold text-slate-900">Double A</div>
          <div className="text-xs text-slate-400">لوحة إدارة المتجر</div>
        </div>
      </div>
      <nav className="space-y-1 p-3">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || pathname.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                active ? "bg-blue-50 text-blue-700" : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
              )}
            >
              <Icon className="h-4 w-4" />
              {label}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
