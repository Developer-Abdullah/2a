import Link from "next/link";
import { getLocale, getT } from "@/lib/locale-server";

export interface PolicySection {
  heading: string;
  body: string[];
}

// PolicyPage renders a simple titled document (terms, refund, privacy), themed + direction-aware.
export async function PolicyPage({ title, updated, sections }: { title: string; updated: string; sections: PolicySection[] }) {
  const [locale, t] = await Promise.all([getLocale(), getT()]);
  return (
    <div className="container-page max-w-3xl py-12">
      <nav className="mb-6 text-sm text-muted-foreground">
        <Link href="/" className="hover:text-brand-600">{t("product.home")}</Link> / <span className="text-foreground/80">{title}</span>
      </nav>
      <h1 className="text-3xl font-extrabold text-foreground">{title}</h1>
      <p className="mt-2 text-sm text-muted-foreground">{locale === "ar" ? "آخر تحديث" : "Last updated"}: {updated}</p>
      <div className="mt-8 space-y-8">
        {sections.map((s, i) => (
          <section key={i}>
            <h2 className="text-lg font-extrabold text-foreground">{s.heading}</h2>
            <div className="mt-2 space-y-2 text-sm leading-8 text-muted-foreground">
              {s.body.map((p, j) => (
                <p key={j}>{p}</p>
              ))}
            </div>
          </section>
        ))}
      </div>
    </div>
  );
}
