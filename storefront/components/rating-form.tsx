"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Star, Loader2 } from "lucide-react";

// RatingForm lets a visitor submit a 1-5 star rating with an optional comment for a product.
export function RatingForm({ slug }: { slug: string }) {
  const router = useRouter();
  const [rating, setRating] = useState(0);
  const [hover, setHover] = useState(0);
  const [comment, setComment] = useState("");
  const [pending, setPending] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (rating < 1) {
      setError("اختر عدد النجوم أولًا");
      return;
    }
    setError("");
    setPending(true);
    try {
      const res = await fetch("/api/rate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ slug, rating, comment }),
      });
      if (res.ok) {
        setDone(true);
        router.refresh();
      } else if (res.status === 429) {
        setError("لقد قيّمت من قبل. حاول لاحقًا.");
      } else {
        setError("تعذّر إرسال التقييم. حاول مرة أخرى.");
      }
    } catch {
      setError("تعذّر الاتصال. حاول مرة أخرى.");
    } finally {
      setPending(false);
    }
  }

  if (done) {
    return <p className="rounded-xl bg-emerald-50 p-4 text-center text-sm font-bold text-emerald-700">شكرًا لتقييمك! 🌟</p>;
  }

  return (
    <form onSubmit={submit} className="rounded-2xl bg-slate-100 p-5">
      <h3 className="font-extrabold text-ink">قيّم تجربتك</h3>
      <div className="mt-3 flex items-center gap-1">
        {Array.from({ length: 5 }).map((_, i) => {
          const val = i + 1;
          const active = val <= (hover || rating);
          return (
            <button
              key={val}
              type="button"
              onClick={() => setRating(val)}
              onMouseEnter={() => setHover(val)}
              onMouseLeave={() => setHover(0)}
              aria-label={`${val} نجوم`}
              className="p-1"
            >
              <Star className={active ? "h-8 w-8 fill-amber-400 text-amber-400" : "h-8 w-8 text-slate-300"} />
            </button>
          );
        })}
      </div>
      <textarea
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        maxLength={280}
        rows={3}
        placeholder="اكتب رأيك (اختياري)…"
        className="mt-3 w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
      />
      {error ? <p className="mt-2 text-sm text-rose-600">{error}</p> : null}
      <button type="submit" disabled={pending} className="btn-primary mt-3">
        {pending ? <Loader2 className="h-5 w-5 animate-spin" /> : <Star className="h-5 w-5" />}
        إرسال التقييم
      </button>
    </form>
  );
}
