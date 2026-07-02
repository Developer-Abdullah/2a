import Link from "next/link";
import { ShieldCheck, Zap, Headphones, BadgeCheck, Star, Package } from "lucide-react";
import { fetchProducts, getCurrency } from "@/lib/api";
import { ProductCard } from "@/components/product-card";
import { Stars } from "@/components/stars";

export const dynamic = "force-dynamic";

const FEATURES = [
  { icon: Zap, title: "تفعيل فوري", body: "يصلك الكود مباشرة بعد الدفع وتفعّله خلال دقائق." },
  { icon: ShieldCheck, title: "بدون جيلبريك", body: "تثبيت آمن بدون أي مخاطرة على جهازك." },
  { icon: BadgeCheck, title: "ضمان حقيقي", body: "ضمان استبدال على كل عملية شراء." },
  { icon: Headphones, title: "دعم 24/7", body: "فريق دعم فني جاهز لمساعدتك في أي وقت." },
];

const REVIEWS = [
  { name: "أحمد", stars: 5, text: "تفعيل سريع جدًا والكود اشتغل من أول مرة. تجربة ممتازة!" },
  { name: "سارة", stars: 5, text: "أفضل سعر لقيته والدعم رد عليّ بسرعة. أنصح فيه." },
  { name: "خالد", stars: 4, text: "خدمة رائعة وسهلة، هجدد الاشتراك أكيد السنة الجاية." },
];

export default async function HomePage() {
  const currency = await getCurrency();
  const products = await fetchProducts(currency);

  return (
    <>
      {/* Hero */}
      <section className="brand-gradient text-white">
        <div className="container-page grid gap-8 py-16 md:grid-cols-2 md:items-center md:py-24">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full bg-white/15 px-4 py-1.5 text-sm font-bold">
              <Package className="h-4 w-4" /> أكثر من 9000 تطبيق بلس
            </span>
            <h1 className="mt-5 text-4xl font-extrabold leading-tight md:text-5xl">
              اشتراك تطبيقات بلس بتفعيل فوري
            </h1>
            <p className="mt-4 max-w-lg text-lg text-white/90">
              كود واحد يفتح لك مكتبة ضخمة من التطبيقات المعدّلة للآيفون والآيباد — بدون جيلبريك، مع ضمان ودعم على مدار الساعة.
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Link href="#products" className="btn-white text-base">تصفّح الباقات</Link>
              <Link href="#features" className="btn bg-white/15 text-white hover:bg-white/25 text-base">لماذا نحن؟</Link>
            </div>
            <div className="mt-8 flex items-center gap-3 text-sm text-white/90">
              <Stars value={4.99} />
              <span className="font-bold">4.99 / 5</span>
              <span>· آلاف العملاء السعداء</span>
            </div>
          </div>
          <div className="hidden justify-self-center md:block">
            <img
              src="/logo.jpg"
              alt="Double A"
              className="h-64 w-64 rounded-3xl shadow-2xl ring-1 ring-white/25"
            />
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="container-page py-16">
        <h2 className="text-center text-3xl font-extrabold text-ink">لماذا تختار Double A؟</h2>
        <div className="mt-10 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {FEATURES.map((f) => (
            <div key={f.title} className="card p-6 text-center">
              <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-50 text-brand-600">
                <f.icon className="h-7 w-7" />
              </div>
              <h3 className="mt-4 text-lg font-extrabold text-ink">{f.title}</h3>
              <p className="mt-2 text-sm leading-6 text-slate-500">{f.body}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Products */}
      <section id="products" className="container-page py-8">
        <h2 className="text-center text-3xl font-extrabold text-ink">اختر باقتك</h2>
        <p className="mt-2 text-center text-slate-500">أسعار تنافسية وتفعيل فوري لكل الأجهزة.</p>
        {products.length === 0 ? (
          <p className="mt-10 text-center text-slate-400">لا توجد منتجات متاحة حاليًا.</p>
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
        <h2 className="text-center text-3xl font-extrabold text-ink">آراء عملائنا</h2>
        <div className="mt-10 grid gap-5 md:grid-cols-3">
          {REVIEWS.map((r) => (
            <div key={r.name} className="card p-6">
              <Stars value={r.stars} />
              <p className="mt-3 text-sm leading-7 text-slate-600">{r.text}</p>
              <div className="mt-4 flex items-center gap-2 text-sm font-bold text-ink">
                <span className="flex h-9 w-9 items-center justify-center rounded-full bg-brand-100 text-brand-700">
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
          <h2 className="text-2xl font-extrabold md:text-3xl">جاهز تبدأ؟ فعّل اشتراكك خلال دقائق</h2>
          <Link href="#products" className="btn-white text-base">اطلب الآن</Link>
        </div>
      </section>
    </>
  );
}
