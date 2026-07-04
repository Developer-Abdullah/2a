import Link from "next/link";
import { ShieldCheck, Zap, Headphones, BadgeCheck, Star, Package } from "lucide-react";
import { fetchProducts, getCurrency } from "@/lib/api";
import { getLocale, getT } from "@/lib/locale-server";
import { ProductCard } from "@/components/product-card";
import { Stars } from "@/components/stars";

export const dynamic = "force-dynamic";

const REVIEWS = {
  ar: [
    { name: "أحمد", stars: 5, text: "تفعيل سريع جدًا والكود اشتغل من أول مرة. تجربة ممتازة!" },
    { name: "سارة", stars: 5, text: "أفضل سعر لقيته والدعم رد عليّ بسرعة. أنصح فيه." },
    { name: "خالد", stars: 4, text: "خدمة رائعة وسهلة، هجدد الاشتراك أكيد السنة الجاية." },
  ],
  en: [
    { name: "Ahmed", stars: 5, text: "Super fast activation and the code worked first try. Excellent!" },
    { name: "Sara", stars: 5, text: "Best price I found and support replied quickly. Recommended." },
    { name: "Khaled", stars: 4, text: "Great, easy service — I'll renew next year for sure." },
  ],
};

export default async function HomePage() {
  const [currency, locale, t] = await Promise.all([getCurrency(), getLocale(), getT()]);
  const products = await fetchProducts(currency);

  const features = [
    { icon: Zap, title: t("feat.instant.t"), body: t("feat.instant.b") },
    { icon: ShieldCheck, title: t("feat.nojb.t"), body: t("feat.nojb.b") },
    { icon: BadgeCheck, title: t("feat.warranty.t"), body: t("feat.warranty.b") },
    { icon: Headphones, title: t("feat.support.t"), body: t("feat.support.b") },
  ];
  const reviews = REVIEWS[locale];

  return (
    <>
      {/* Hero */}
      <section className="brand-gradient text-white">
        <div className="container-page grid gap-8 py-16 md:grid-cols-2 md:items-center md:py-24">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full bg-white/15 px-4 py-1.5 text-sm font-bold">
              <Package className="h-4 w-4" /> {t("home.badge")}
            </span>
            <h1 className="mt-5 text-3xl font-extrabold leading-tight sm:text-4xl md:text-5xl">{t("home.title")}</h1>
            <p className="mt-4 max-w-lg text-base text-white/90 sm:text-lg">{t("home.subtitle")}</p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Link href="#products" className="btn-white text-base">{t("home.browse")}</Link>
              <Link href="#features" className="btn bg-white/15 text-base text-white hover:bg-white/25">{t("home.why")}</Link>
            </div>
            <div className="mt-8 flex flex-wrap items-center gap-3 text-sm text-white/90">
              <Stars value={4.99} />
              <span className="font-bold">4.99 / 5</span>
              <span>{t("home.happy")}</span>
            </div>
          </div>
          <div className="hidden justify-self-center md:block">
            <img src="/logo.jpg" alt="Double A" className="h-64 w-64 rounded-3xl shadow-2xl ring-1 ring-white/25" />
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="container-page py-16">
        <h2 className="text-center text-2xl font-extrabold text-foreground sm:text-3xl">{t("home.features_title")}</h2>
        <div className="mt-10 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {features.map((f) => (
            <div key={f.title} className="card p-6 text-center">
              <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-50 text-brand-600 dark:bg-brand-800/40 dark:text-brand-200">
                <f.icon className="h-7 w-7" />
              </div>
              <h3 className="mt-4 text-lg font-extrabold text-foreground">{f.title}</h3>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">{f.body}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Products */}
      <section id="products" className="container-page py-8">
        <h2 className="text-center text-2xl font-extrabold text-foreground sm:text-3xl">{t("home.products_title")}</h2>
        <p className="mt-2 text-center text-muted-foreground">{t("home.products_sub")}</p>
        {products.length === 0 ? (
          <p className="mt-10 text-center text-muted-foreground">{t("home.no_products")}</p>
        ) : (
          <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {products.map((p) => (
              <ProductCard key={p.id} product={p} />
            ))}
          </div>
        )}
      </section>

      {/* Reviews */}
      <section id="reviews" className="container-page py-16">
        <h2 className="text-center text-2xl font-extrabold text-foreground sm:text-3xl">{t("home.reviews_title")}</h2>
        <div className="mt-10 grid gap-5 md:grid-cols-3">
          {reviews.map((r) => (
            <div key={r.name} className="card p-6">
              <Stars value={r.stars} />
              <p className="mt-3 text-sm leading-7 text-muted-foreground">{r.text}</p>
              <div className="mt-4 flex items-center gap-2 text-sm font-bold text-foreground">
                <span className="flex h-9 w-9 items-center justify-center rounded-full bg-brand-100 text-brand-700 dark:bg-brand-800/60 dark:text-brand-100">
                  {r.name.charAt(0)}
                </span>
                {r.name}
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="container-page pb-4">
        <div className="brand-gradient flex flex-col items-center gap-4 rounded-3xl px-6 py-12 text-center text-white">
          <Star className="h-10 w-10" />
          <h2 className="text-2xl font-extrabold md:text-3xl">{t("home.cta_title")}</h2>
          <Link href="#products" className="btn-white text-base">{t("home.cta_btn")}</Link>
        </div>
      </section>
    </>
  );
}
