"use client";

import { useState } from "react";
import { Check, Copy, KeyRound } from "lucide-react";

export function OrderCodes({ codes }: { codes: string[] }) {
  if (!codes || codes.length === 0) return null;
  return (
    <div className="rounded-2xl border-2 border-dashed border-brand-300 bg-brand-50 p-5">
      <h2 className="flex items-center gap-2 font-extrabold text-brand-800">
        <KeyRound className="h-5 w-5" /> {codes.length > 1 ? "أكواد التفعيل" : "كود التفعيل"}
      </h2>
      <div className="mt-4 space-y-3">
        {codes.map((code) => (
          <CodeRow key={code} code={code} />
        ))}
      </div>
      <p className="mt-4 text-xs text-brand-700/80">
        احتفظ بالكود في مكان آمن. صالح لعملية تفعيل واحدة.
      </p>
    </div>
  );
}

function CodeRow({ code }: { code: string }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1600);
    } catch {
      /* clipboard unavailable */
    }
  };
  return (
    <div className="flex items-center justify-between gap-3 rounded-xl bg-white p-3 ring-1 ring-brand-100">
      <code className="select-all font-mono text-lg font-bold tracking-wider text-ink">{code}</code>
      <button onClick={copy} className="btn-ghost px-3 py-2 text-sm">
        {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
        {copied ? "تم النسخ" : "نسخ"}
      </button>
    </div>
  );
}
