"use client";

import { useState } from "react";
import Link from "next/link";
import { Search, Loader2, PackageSearch } from "lucide-react";
import { formatMoney } from "@/lib/format";
import { useStore } from "@/components/store-provider";

interface OrderRow {
  id: string;
  email: string;
  currency: string;
  total: number;
  status: string;
  created_at: string;
  code_count: number;
}

const STATUS_CLS: Record<string, string> = {
  pending: "bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300",
  paid: "bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300",
  failed: "bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300",
  fulfilled: "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300",
};

export default function MyOrdersPage() {
  const { t } = useStore();
  const [email, setEmail] = useState("");
  const [orders, setOrders] = useState<OrderRow[] | null>(null);
  const [loading, setLoading] = useState(false);

  async function lookup(e: React.FormEvent) {
    e.preventDefault();
    if (!email) return;
    setLoading(true);
    try {
      const res = await fetch(`/api/my-orders?email=${encodeURIComponent(email)}`);
      const data = await res.json();
      setOrders(data.orders ?? []);
    } catch {
      setOrders([]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="container-page max-w-2xl py-12">
      <h1 className="text-2xl font-extrabold text-foreground">{t("myorders.title")}</h1>
      <p className="mt-2 text-muted-foreground">{t("myorders.sub")}</p>

      <form onSubmit={lookup} className="mt-6 flex gap-2">
        <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} placeholder="name@example.com" className="field flex-1" />
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? <Loader2 className="h-5 w-5 animate-spin" /> : <Search className="h-5 w-5" />}
          {t("myorders.search")}
        </button>
      </form>

      {orders !== null && (
        <div className="mt-8">
          {orders.length === 0 ? (
            <div className="card p-8 text-center text-muted-foreground">
              <PackageSearch className="mx-auto h-12 w-12 text-muted-foreground/40" />
              <p className="mt-3">{t("myorders.none")}</p>
            </div>
          ) : (
            <ul className="space-y-3">
              {orders.map((o) => (
                <li key={o.id}>
                  <Link href={`/order/${o.id}`} className="card flex items-center justify-between gap-3 p-4 transition hover:shadow-lg">
                    <div className="min-w-0">
                      <div className="font-mono text-xs text-muted-foreground">#{o.id.slice(0, 8)}</div>
                      <div className="mt-1 font-bold text-foreground">{formatMoney(o.total, o.currency)}</div>
                      <div className="text-xs text-muted-foreground">{new Date(o.created_at).toLocaleString()}</div>
                    </div>
                    <div className="flex flex-none items-center gap-3">
                      {o.code_count > 0 && <span className="text-xs text-muted-foreground">{o.code_count} {t("myorders.codes")}</span>}
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${STATUS_CLS[o.status] || "bg-muted text-muted-foreground"}`}>
                        {t(`status.${o.status}`)}
                      </span>
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
