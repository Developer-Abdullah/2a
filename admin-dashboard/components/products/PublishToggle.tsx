"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { setProductPublishedAction } from "@/lib/actions";
import { Button } from "@/components/ui/button";

// PublishToggle flips a product's storefront visibility from the edit page.
export default function PublishToggle({ id, published }: { id: string; published: boolean }) {
  const router = useRouter();
  const [state, setState] = useState(published);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function toggle() {
    setPending(true);
    setError("");
    const res = await setProductPublishedAction(id, !state);
    setPending(false);
    if (res.ok) {
      setState(!state);
      router.refresh();
    } else {
      setError(res.error || "تعذّر التحديث");
    }
  }

  return (
    <div className="space-y-2 border-t border-slate-100 pt-3">
      <div className="flex items-center justify-between">
        <span>الحالة</span>
        <span className={state ? "font-bold text-green-600" : "font-bold text-slate-500"}>
          {state ? "منشور" : "مسودة"}
        </span>
      </div>
      <Button type="button" variant={state ? "outline" : "default"} onClick={toggle} disabled={pending} className="w-full">
        {pending ? "…" : state ? "إلغاء النشر" : "نشر في المتجر"}
      </Button>
      {error ? <p className="text-xs text-red-600">{error}</p> : null}
    </div>
  );
}
