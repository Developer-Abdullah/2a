"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Star, Loader2 } from "lucide-react";
import { useStore } from "./store-provider";

// RatingForm lets a visitor submit a 1-5 star rating with an optional comment for a product.
export function RatingForm({ slug }: { slug: string }) {
  const router = useRouter();
  const { t } = useStore();
  const [rating, setRating] = useState(0);
  const [hover, setHover] = useState(0);
  const [comment, setComment] = useState("");
  const [pending, setPending] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (rating < 1) {
      setError(t("rate.pick_first"));
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
        setError(t("rate.err_rated"));
      } else {
        setError(t("rate.err_generic"));
      }
    } catch {
      setError(t("rate.err_conn"));
    } finally {
      setPending(false);
    }
  }

  if (done) {
    return <p className="rounded-xl bg-emerald-50 p-4 text-center text-sm font-bold text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">{t("rate.thanks")}</p>;
  }

  return (
    <form onSubmit={submit} className="rounded-2xl bg-muted p-5">
      <h3 className="font-extrabold text-foreground">{t("rate.title")}</h3>
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
              aria-label={`${val}`}
              className="p-1"
            >
              <Star className={active ? "h-8 w-8 fill-amber-400 text-amber-400" : "h-8 w-8 text-muted-foreground/40"} />
            </button>
          );
        })}
      </div>
      <textarea
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        maxLength={280}
        rows={3}
        placeholder={t("rate.placeholder")}
        className="field mt-3 text-sm"
      />
      {error ? <p className="mt-2 text-sm text-rose-600 dark:text-rose-400">{error}</p> : null}
      <button type="submit" disabled={pending} className="btn-primary mt-3">
        {pending ? <Loader2 className="h-5 w-5 animate-spin" /> : <Star className="h-5 w-5" />}
        {t("rate.submit")}
      </button>
    </form>
  );
}
