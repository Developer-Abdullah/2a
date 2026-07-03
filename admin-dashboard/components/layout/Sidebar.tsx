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
import { useI18n } from "@/components/providers/LocaleProvider";

// Single-store admin nav. Labels are i18n keys resolved at render so the sidebar follows the active
// language. The storefront catalog + orders are the primary surfaces.
const storeNav = [
  { href: "/dashboard", key: "nav.dashboard", icon: LayoutDashboard },
  { href: "/products", key: "nav.products", icon: Package },
  { href: "/orders", key: "nav.orders", icon: ShoppingCart },
  { href: "/apps", key: "nav.apps", icon: AppWindow },
  { href: "/activation", key: "nav.codes", icon: KeyRound },
  { href: "/ratings", key: "nav.ratings", icon: Star },
  { href: "/notifications", key: "nav.notifications", icon: Bell },
  { href: "/users", key: "nav.customers", icon: Users },
  { href: "/devices", key: "nav.devices", icon: Smartphone },
  { href: "/analytics", key: "nav.analytics", icon: BarChart3 },
  { href: "/enrollment", key: "nav.enrollment", icon: ShieldCheck },
  { href: "/settings", key: "nav.settings", icon: Settings },
];

// isOwner is accepted for backward compatibility but no longer switches the nav — there is one store.
export default function Sidebar({ isOwner: _isOwner = false }: { isOwner?: boolean }) {
  const pathname = usePathname();
  const { t } = useI18n();

  return (
    <aside className="w-64 shrink-0 border-e border-border bg-card">
      <div className="flex h-16 items-center gap-3 border-b border-border bg-gradient-to-l from-brand-900 to-accent-700 px-6">
        <img src="/logo.jpg" alt="Double A" className="h-9 w-9 rounded-lg ring-1 ring-white/30" />
        <div className="leading-tight">
          <div className="text-sm font-bold text-white">Double A</div>
          <div className="text-xs text-white/70">{t("brand.subtitle")}</div>
        </div>
      </div>
      <nav className="space-y-1 p-3">
        {storeNav.map(({ href, key, icon: Icon }) => {
          const active = pathname === href || pathname.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                active
                  ? "bg-brand-50 text-brand-800 dark:bg-brand-800/40 dark:text-brand-100"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground"
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {t(key)}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
