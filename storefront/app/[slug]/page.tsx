import { notFound } from "next/navigation";
import Link from "next/link";
import { Check, FileText, ShoppingBag } from "lucide-react";
import { fetchProduct } from "@/lib/api";
import { BuyBox } from "@/components/buy-box";
import { Stars } from "@/components/stars";

export const dynamic = "force-dynamic";

export default async function ProductPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const product = await fetchProduct(slug);
  if (!product) notFound();

  const deviceLabel =
    product.device_type === "ipad" ? "آيباد" : product.device_type === "iphone" ? "آيفون" : "آيفون / آيباد";

  return (
    <div className="container-page py-8">
      <nav className="mb-6 flex items-center gap-2 text-sm text-slate-400">
        <Link href="/" className="hover:text-brand-700">الرئيسية</Link>
        <span>/</span>
        <span className="text-slate-600">{product.name}</span>
      </nav>

      <div className="grid gap-8 lg:grid-cols-3">
        {/* Main */}
        <div className="lg:col-span-2">
          <div className="brand-gradient flex h-44 items-center justify-center rounded-3xl text-white">
            <div className="text-center">
              <ShoppingBag className="mx-auto h-12 w-12" />
              <p className="mt-2 text-lg font-extrabold">{deviceLabel} · تفعيل فوري</p>
            </div>
          </div>

          <div className="mt-6 flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-extrabold text-ink md:text-3xl">{product.name}</h1>
            <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-700">
              تم شراؤه {product.purchase_count.toLocaleString("ar-EG")} مرة
            </span>
          </div>

          <div className="mt-3 flex items-center gap-2 text-sm text-slate-500">
            <Stars value={product.rating_avg} />
            <span className="font-bold text-ink">{product.rating_avg.toFixed(2)}</span>
            <span>({product.rating_count} تقييم)</span>
          </div>

          {product.description && (
            <p className="mt-5 leading-8 text-slate-600">{product.description}</p>
          )}

          {product.features.length > 0 && (
            <div className="mt-8">
              <h2 className="text-lg font-extrabold text-ink">المميزات</h2>
              <ul className="mt-4 grid gap-3 sm:grid-cols-2">
                {product.features.map((f, i) => (
                  <li key={i} className="flex items-start gap-2 text-slate-600">
                    <span className="mt-0.5 flex h-6 w-6 flex-none items-center justify-center rounded-full bg-brand-50 text-brand-600">
                      <Check className="h-4 w-4" />
                    </span>
                    {f}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {product.video_url && (
            <div className="mt-8">
              <h2 className="text-lg font-extrabold text-ink">شرح التفعيل</h2>
              <div className="mt-4 aspect-video overflow-hidden rounded-2xl ring-1 ring-slate-200">
                <iframe src={product.video_url} className="h-full w-full" allowFullScreen title="شرح التفعيل" />
              </div>
            </div>
          )}

          {product.terms.length > 0 && (
            <div className="mt-8 rounded-2xl bg-slate-100 p-6">
              <h2 className="flex items-center gap-2 text-lg font-extrabold text-ink">
                <FileText className="h-5 w-5 text-slate-500" /> الشروط والأحكام
              </h2>
              <ul className="mt-3 space-y-2 text-sm leading-7 text-slate-600">
                {product.terms.map((t, i) => (
                  <li key={i} className="flex gap-2"><span className="text-brand-500">•</span>{t}</li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Buy box */}
        <div className="lg:col-span-1">
          <div className="lg:sticky lg:top-24">
            <BuyBox product={product} />
          </div>
        </div>
      </div>
    </div>
  );
}
