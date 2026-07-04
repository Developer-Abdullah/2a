"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Wallet, MessageCircle, Copy, Check, Upload, Loader2, CheckCircle2 } from "lucide-react";
import { PaymentMethod } from "@/lib/payment";
import { useStore } from "./store-provider";

// PaymentPanel shows manual-payment instructions for a pending order: copyable transfer numbers, a
// WhatsApp button, and a transfer-screenshot uploader. `hasProof` reflects whether one is already up.
export function PaymentPanel({
  orderId,
  total,
  methods,
  whatsapp,
  hasProof,
}: {
  orderId: string;
  total: string;
  methods: PaymentMethod[];
  whatsapp: string;
  hasProof: boolean;
}) {
  const router = useRouter();
  const { t } = useStore();
  const [copied, setCopied] = useState<string>("");
  const [uploaded, setUploaded] = useState(hasProof);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");

  const waText = encodeURIComponent(`${orderId} — ${total}`);
  const waLink = whatsapp ? `https://wa.me/${whatsapp}?text=${waText}` : "";

  async function copy(num: string) {
    try {
      await navigator.clipboard.writeText(num);
      setCopied(num);
      setTimeout(() => setCopied(""), 1600);
    } catch {
      /* clipboard unavailable */
    }
  }

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setError("");
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append("file", file);
      const res = await fetch(`/api/order-proof/${orderId}`, { method: "POST", body: fd });
      if (res.ok) {
        setUploaded(true);
        router.refresh();
      } else if (res.status === 429) {
        setError(t("pay.err_many"));
      } else {
        setError(t("pay.err"));
      }
    } catch {
      setError(t("pay.err_conn"));
    } finally {
      setUploading(false);
    }
  }

  return (
    <div className="rounded-2xl border-2 border-dashed border-amber-300 bg-amber-50 p-5 dark:border-amber-700/60 dark:bg-amber-900/20">
      <h2 className="flex items-center gap-2 font-extrabold text-amber-900 dark:text-amber-200">
        <Wallet className="h-5 w-5" /> {t("pay.title")}
      </h2>
      <p className="mt-2 text-sm leading-7 text-amber-900 dark:text-amber-200/90">{t("pay.transfer_note")}</p>

      <div className="mt-3 rounded-xl bg-card p-3 text-sm ring-1 ring-border">
        <div className="flex items-center justify-between border-b border-border pb-2">
          <span className="text-muted-foreground">{t("pay.amount")}</span>
          <span className="font-extrabold text-foreground">{total}</span>
        </div>
        <ul className="mt-2 space-y-2">
          {methods.map((m) => (
            <li key={m.number} className="flex items-center justify-between gap-2">
              <div>
                <div className="text-xs text-muted-foreground">{m.name}</div>
                <div className="font-mono text-base font-bold tracking-wide text-foreground">{m.number}</div>
              </div>
              <button onClick={() => copy(m.number)} className="btn-ghost px-3 py-2 text-sm">
                {copied === m.number ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                {copied === m.number ? t("pay.copied") : t("pay.copy_num")}
              </button>
            </li>
          ))}
        </ul>
        <div className="mt-2 border-t border-border pt-2 text-xs text-muted-foreground">
          {t("pay.ref")}: <span className="font-mono text-foreground/80">{orderId}</span>
        </div>
      </div>

      {/* Upload transfer screenshot */}
      <div className="mt-4">
        {uploaded ? (
          <div className="flex items-center gap-2 rounded-xl bg-emerald-100 p-3 text-sm font-bold text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
            <CheckCircle2 className="h-5 w-5" /> {t("pay.uploaded")}
          </div>
        ) : (
          <label className="flex cursor-pointer items-center justify-center gap-2 rounded-xl border-2 border-dashed border-amber-400 bg-card px-4 py-3 text-sm font-bold text-amber-800 hover:bg-amber-50 dark:text-amber-300 dark:hover:bg-amber-900/20">
            {uploading ? <Loader2 className="h-5 w-5 animate-spin" /> : <Upload className="h-5 w-5" />}
            {uploading ? t("pay.uploading") : t("pay.upload")}
            <input type="file" accept="image/*" onChange={onFile} disabled={uploading} className="hidden" />
          </label>
        )}
        {error ? <p className="mt-2 text-sm text-rose-600 dark:text-rose-400">{error}</p> : null}
      </div>

      {waLink && (
        <a href={waLink} target="_blank" rel="noopener noreferrer" className="btn-primary mt-3 w-full">
          <MessageCircle className="h-5 w-5" /> {t("pay.whatsapp")}
        </a>
      )}
      <p className="mt-3 text-center text-xs text-amber-700/80 dark:text-amber-300/70">{t("pay.footer_note")}</p>
    </div>
  );
}
