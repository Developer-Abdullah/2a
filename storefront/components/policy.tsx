import Link from "next/link";

export interface PolicySection {
  heading: string;
  body: string[];
}

// PolicyPage renders a simple titled document (terms, refund, privacy) in RTL.
export function PolicyPage({ title, updated, sections }: { title: string; updated: string; sections: PolicySection[] }) {
  return (
    <div className="container-page max-w-3xl py-12">
      <nav className="mb-6 text-sm text-slate-400">
        <Link href="/" className="hover:text-brand-700">الرئيسية</Link> / <span className="text-slate-600">{title}</span>
      </nav>
      <h1 className="text-3xl font-extrabold text-ink">{title}</h1>
      <p className="mt-2 text-sm text-slate-400">آخر تحديث: {updated}</p>
      <div className="mt-8 space-y-8">
        {sections.map((s, i) => (
          <section key={i}>
            <h2 className="text-lg font-extrabold text-ink">{s.heading}</h2>
            <div className="mt-2 space-y-2 text-sm leading-8 text-slate-600">
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
