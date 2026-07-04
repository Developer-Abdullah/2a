import { notFound } from "next/navigation";
import Link from "next/link";
import { Check, FileText, ShoppingBag } from "lucide-react";
import { fetchProduct, fetchReviews } from "@/lib/api";
import { getT } from "@/lib/locale-server";
import { BuyBox } from "@/components/buy-box";
import { Stars } from "@/components/stars";
import { RatingForm } from "@/components/rating-form";

export const dynamic = "force-dynamic";

export default async function ProductPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const t = await getT();
  const product = await fetchProduct(slug);
  if (!product) notFound();
  const reviews = await fetchReviews(slug);

  const deviceLabel =
    product.device_type === "ipad" ? t("device.ipad") : product.device_type === "iphone" ? t("device.iphone") : t("device.both");

  return (
    <div className="container-page py-8">
      <nav className="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/" className="hover:text-brand-600">{t("product.home")}</Link>
        <span>/</span>
        <span className="text-foreground/80">{product.name}</span>
      </nav>

      <div className="grid gap-8 lg:grid-cols-3">
        {/* Main */}
        <div className="lg:col-span-2">
          {product.image_url ? (
            <div className="h-56 overflow-hidden rounded-3xl sm:h-72">
              <img src={`/api/product-image/${product.slug}`} alt={product.name} className="h-full w-full object-cover" />
            </div>
          ) : (
            <div className="brand-gradient flex h-44 items-center justify-center rounded-3xl text-white">
              <div className="text-center">
                <ShoppingBag className="mx-auto h-12 w-12" />
                <p className="mt-2 text-lg font-extrabold">{deviceLabel} · {t("product.instant_label")}</p>
              </div>
            </div>
          )}

          <div className="mt-6 flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-extrabold text-foreground md:text-3xl">{product.name}</h1>
            <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
              {t("product.purchased")} {product.purchase_count.toLocaleString()} {t("product.times")}
            </span>
          </div>

          <div className="mt-3 flex items-center gap-2 text-sm text-muted-foreground">
            <Stars value={product.rating_avg} />
            <span className="font-bold text-foreground">{product.rating_avg.toFixed(2)}</span>
            <span>({product.rating_count})</span>
          </div>

          {product.description && <p className="mt-5 leading-8 text-muted-foreground">{product.description}</p>}

          {product.features.length > 0 && (
            <div className="mt-8">
              <h2 className="text-lg font-extrabold text-foreground">{t("product.features")}</h2>
              <ul className="mt-4 grid gap-3 sm:grid-cols-2">
                {product.features.map((f, i) => (
                  <li key={i} className="flex items-start gap-2 text-muted-foreground">
                    <span className="mt-0.5 flex h-6 w-6 flex-none items-center justify-center rounded-full bg-brand-50 text-brand-600 dark:bg-brand-800/40 dark:text-brand-200">
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
              <h2 className="text-lg font-extrabold text-foreground">{t("product.instant_label")}</h2>
              <div className="mt-4 aspect-video overflow-hidden rounded-2xl ring-1 ring-border">
                <iframe src={product.video_url} className="h-full w-full" allowFullScreen title="video" />
              </div>
            </div>
          )}

          {product.terms.length > 0 && (
            <div className="mt-8 rounded-2xl bg-muted p-6">
              <h2 className="flex items-center gap-2 text-lg font-extrabold text-foreground">
                <FileText className="h-5 w-5 text-muted-foreground" /> {t("product.terms")}
              </h2>
              <ul className="mt-3 space-y-2 text-sm leading-7 text-muted-foreground">
                {product.terms.map((term, i) => (
                  <li key={i} className="flex gap-2"><span className="text-brand-500">•</span>{term}</li>
                ))}
              </ul>
            </div>
          )}

          {/* Reviews */}
          <div className="mt-10">
            <h2 className="text-lg font-extrabold text-foreground">{t("product.reviews")} ({product.rating_count})</h2>
            {reviews.length > 0 ? (
              <ul className="mt-4 space-y-3">
                {reviews.map((r, i) => (
                  <li key={i} className="card p-4">
                    <Stars value={r.rating} />
                    <p className="mt-2 text-sm leading-7 text-muted-foreground">{r.comment}</p>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-3 text-sm text-muted-foreground">{t("product.no_reviews")}</p>
            )}
            <div className="mt-6">
              <RatingForm slug={product.slug} />
            </div>
          </div>
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
