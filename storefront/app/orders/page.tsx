"use client";

import { useState } from "react";
import Link from "next/link";
import { Search, Loader2, PackageSearch } from "lucide-react";
import { formatMoney } from "@/lib/format";

interface OrderRow {
  id: string;
  email: string;
  currency: string;
  total: number;
  status: string;
  created_at: string;
  code_count: number;
}

const STATUS: Record<string, { label: string; cls: string }> = {
  pending: { label: "قيد المراجعة", cls: "bg-amber-100 text-amber-700" },
  paid: { label: "مدفوع", cls: "bg-blue-100 text-blue-700" },
  failed: { label: "ملغى", cls: "bg-rose-100 text-rose-700" },
  fulfilled: { label: "مكتمل", cls: "bg-emerald-100 text-emerald-700" },
};

export default function MyOrdersPage() {
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
      <h1 className="text-2xl font-extrabold text-ink">تتبّع طلباتي</h1>
      <p className="mt-2 text-slate-500">أدخل بريدك الإلكتروني لعرض طلباتك وأكوادك.</p>

      <form onSubmit={lookup} className="mt-6 flex gap-2">
        <input
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="name@example.com"
          className="flex-1 rounded-xl border border-slate-200 bg-white px-4 py-3 outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
        />
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? <Loader2 className="h-5 w-5 animate-spin" /> : <Search className="h-5 w-5" />}
          بحث
        </button>
      </form>

      {orders !== null && (
        <div className="mt-8">
          {orders.length === 0 ? (
            <div className="card p-8 text-center text-slate-500">
              <PackageSearch className="mx-auto h-12 w-12 text-slate-300" />
              <p className="mt-3">لا توجد طلبات مرتبطة بهذا البريد.</p>
            </div>
          ) : (
            <ul className="space-y-3">
              {orders.map((o) => {
                const s = STATUS[o.status] ?? { label: o.status, cls: "bg-slate-100 text-slate-600" };
                return (
                  <li key={o.id}>
                    <Link href={`/order/${o.id}`} className="card flex items-center justify-between gap-3 p-4 transition hover:shadow-lg">
                      <div>
                        <div className="font-mono text-xs text-slate-400">#{o.id.slice(0, 8)}</div>
                        <div className="mt-1 font-bold text-ink">{formatMoney(o.total, o.currency)}</div>
                        <div className="text-xs text-slate-400">{new Date(o.created_at).toLocaleString("ar-EG")}</div>
                      </div>
                      <div className="flex items-center gap-3">
                        {o.code_count > 0 && <span className="text-xs text-slate-500">{o.code_count} كود</span>}
                        <span className={`rounded-full px-3 py-1 text-xs font-bold ${s.cls}`}>{s.label}</span>
                      </div>
                    </Link>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
