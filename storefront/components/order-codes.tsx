"use client";

import { useState } from "react";
import { Check, Copy, KeyRound } from "lucide-react";
import { useStore } from "./store-provider";

export function OrderCodes({ codes }: { codes: string[] }) {
  const { t } = useStore();
  if (!codes || codes.length === 0) return null;
  return (
    <div className="rounded-2xl border-2 border-dashed border-brand-300 bg-brand-50 p-5 dark:border-brand-700 dark:bg-brand-900/40">
      <h2 className="flex items-center gap-2 font-extrabold text-brand-800 dark:text-brand-100">
        <KeyRound className="h-5 w-5" /> {t("order.your_codes")}
      </h2>
      <div className="mt-4 space-y-3">
        {codes.map((code) => (
          <CodeRow key={code} code={code} copy={t("order.copy")} copied={t("order.copied")} />
        ))}
      </div>
    </div>
  );
}

function CodeRow({ code, copy, copied }: { code: string; copy: string; copied: string }) {
  const [done, setDone] = useState(false);
  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setDone(true);
      setTimeout(() => setDone(false), 1600);
    } catch {
      /* clipboard unavailable */
    }
  };
  return (
    <div className="flex items-center justify-between gap-3 rounded-xl bg-card p-3 ring-1 ring-border">
      <code className="select-all font-mono text-lg font-bold tracking-wider text-foreground">{code}</code>
      <button onClick={onCopy} className="btn-ghost px-3 py-2 text-sm">
        {done ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
        {done ? copied : copy}
      </button>
    </div>
  );
}
