import Link from "next/link";
import { Product } from "@/lib/types";
import { Price } from "./price";
import { Stars } from "./stars";
import { getT } from "@/lib/locale-server";
import { Smartphone, Tablet, MonitorSmartphone } from "lucide-react";

function DeviceIcon({ type }: { type: string }) {
  if (type === "iphone") return <Smartphone className="h-5 w-5" />;
  if (type === "ipad") return <Tablet className="h-5 w-5" />;
  return <MonitorSmartphone className="h-5 w-5" />;
}

export async function ProductCard({ product }: { product: Product }) {
  const t = await getT();
  const deviceLabel =
    product.device_type === "ipad" ? t("device.ipad") : product.device_type === "iphone" ? t("device.iphone") : t("device.both");

  return (
    <Link href={`/${product.slug}`} className="card group flex flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-lg">
      {product.image_url ? (
        <div className="relative h-40 overflow-hidden">
          <img src={`/api/product-image/${product.slug}`} alt={product.name} className="h-full w-full object-cover" />
          <div className="absolute top-2 start-2 rounded-xl bg-amber-400 px-3 py-1 text-xs font-extrabold text-ink">{t("card.instant")}</div>
          <div className="absolute top-2 end-2 flex items-center gap-1 rounded-xl bg-black/45 px-2.5 py-1 text-xs font-bold text-white">
            <DeviceIcon type={product.device_type} />
            <span>{deviceLabel}</span>
          </div>
        </div>
      ) : (
        <div className="brand-gradient flex h-28 items-center justify-between px-5 text-white">
          <div className="flex items-center gap-2 rounded-xl bg-white/15 px-3 py-1.5 text-sm font-bold">
            <DeviceIcon type={product.device_type} />
            <span>{deviceLabel}</span>
          </div>
          <div className="rounded-xl bg-amber-400 px-3 py-1 text-xs font-extrabold text-ink">{t("card.instant")}</div>
        </div>
      )}
      <div className="flex flex-1 flex-col gap-3 p-5">
        <h3 className="text-lg font-extrabold leading-7 text-foreground group-hover:text-brand-600">{product.name}</h3>
        <p className="line-clamp-2 text-sm text-muted-foreground">{product.subtitle || product.description}</p>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Stars value={product.rating_avg} />
          <span className="font-bold text-foreground">{product.rating_avg.toFixed(2)}</span>
          <span>({product.rating_count})</span>
        </div>
        <div className="mt-auto flex items-center justify-between pt-2">
          <Price product={product} />
          <span className="btn-ghost px-4 py-2 text-sm">{t("card.details")}</span>
        </div>
      </div>
    </Link>
  );
}
