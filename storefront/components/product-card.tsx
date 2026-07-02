import Link from "next/link";
import { Product } from "@/lib/types";
import { Price } from "./price";
import { Stars } from "./stars";
import { Smartphone, Tablet, MonitorSmartphone } from "lucide-react";

function DeviceIcon({ type }: { type: string }) {
  if (type === "iphone") return <Smartphone className="h-5 w-5" />;
  if (type === "ipad") return <Tablet className="h-5 w-5" />;
  return <MonitorSmartphone className="h-5 w-5" />;
}

export function ProductCard({ product }: { product: Product }) {
  return (
    <Link href={`/${product.slug}`} className="card group flex flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-lg">
      <div className="brand-gradient flex h-28 items-center justify-between px-5 text-white">
        <div className="flex items-center gap-2 rounded-xl bg-white/15 px-3 py-1.5 text-sm font-bold">
          <DeviceIcon type={product.device_type} />
          <span>{product.device_type === "ipad" ? "آيباد" : product.device_type === "iphone" ? "آيفون" : "آيفون / آيباد"}</span>
        </div>
        <div className="rounded-xl bg-amber-400 px-3 py-1 text-xs font-extrabold text-ink">تفعيل فوري</div>
      </div>
      <div className="flex flex-1 flex-col gap-3 p-5">
        <h3 className="text-lg font-extrabold leading-7 text-ink group-hover:text-brand-700">{product.name}</h3>
        <p className="line-clamp-2 text-sm text-slate-500">{product.subtitle || product.description}</p>
        <div className="flex items-center gap-2 text-sm text-slate-500">
          <Stars value={product.rating_avg} />
          <span className="font-bold text-ink">{product.rating_avg.toFixed(2)}</span>
          <span>({product.rating_count})</span>
        </div>
        <div className="mt-auto flex items-center justify-between pt-2">
          <Price product={product} />
          <span className="btn-ghost px-4 py-2 text-sm">التفاصيل</span>
        </div>
      </div>
    </Link>
  );
}
