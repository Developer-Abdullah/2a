"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Wallet, MessageCircle, Copy, Check, Upload, Loader2, CheckCircle2 } from "lucide-react";
import { PaymentMethod } from "@/lib/payment";

// PaymentPanel shows manual-payment instructions for a pending order: copyable transfer numbers, a
// WhatsApp button, and a transfer-screenshot uploader. `hasProof` reflects whether one is already up.
export function PaymentPanel({
  orderId,
  total,
  methods,
  note,
  whatsapp,
  hasProof,
}: {
  orderId: string;
  total: string;
  methods: PaymentMethod[];
  note: string;
  whatsapp: string;
  hasProof: boolean;
}) {
  const router = useRouter();
  const [copied, setCopied] = useState<string>("");
  const [uploaded, setUploaded] = useState(hasProof);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");

  const waText = encodeURIComponent(`مرحبًا، أودّ تأكيد دفع الطلب رقم ${orderId} بقيمة ${total}.`);
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
        setError("محاولات كثيرة. انتظر قليلًا وحاول مجددًا.");
      } else {
        setError("تعذّر رفع الصورة. تأكد أنها صورة وحاول مرة أخرى.");
      }
    } catch {
      setError("تعذّر الاتصال. حاول مرة أخرى.");
    } finally {
      setUploading(false);
    }
  }

  return (
    <div className="rounded-2xl border-2 border-dashed border-amber-300 bg-amber-50 p-5">
      <h2 className="flex items-center gap-2 font-extrabold text-amber-900">
        <Wallet className="h-5 w-5" /> تعليمات الدفع
      </h2>
      <p className="mt-2 text-sm leading-7 text-amber-900">{note}</p>

      <div className="mt-3 rounded-xl bg-white p-3 text-sm ring-1 ring-amber-100">
        <div className="flex items-center justify-between border-b border-slate-100 pb-2">
          <span className="text-slate-500">المبلغ</span>
          <span className="font-extrabold text-ink">{total}</span>
        </div>
        <ul className="mt-2 space-y-2">
          {methods.map((m) => (
            <li key={m.number} className="flex items-center justify-between gap-2">
              <div>
                <div className="text-xs text-slate-400">{m.name}</div>
                <div className="font-mono text-base font-bold tracking-wide text-ink">{m.number}</div>
              </div>
              <button onClick={() => copy(m.number)} className="btn-ghost px-3 py-2 text-sm">
                {copied === m.number ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                {copied === m.number ? "تم النسخ" : "نسخ الرقم"}
              </button>
            </li>
          ))}
        </ul>
        <div className="mt-2 border-t border-slate-100 pt-2 text-xs text-slate-400">
          رقم الطلب (المرجع): <span className="font-mono text-slate-600">{orderId}</span>
        </div>
      </div>

      {/* Upload transfer screenshot */}
      <div className="mt-4">
        {uploaded ? (
          <div className="flex items-center gap-2 rounded-xl bg-emerald-100 p-3 text-sm font-bold text-emerald-700">
            <CheckCircle2 className="h-5 w-5" /> تم استلام إثبات التحويل — بانتظار المراجعة
          </div>
        ) : (
          <label className="flex cursor-pointer items-center justify-center gap-2 rounded-xl border-2 border-dashed border-amber-400 bg-white px-4 py-3 text-sm font-bold text-amber-800 hover:bg-amber-100">
            {uploading ? <Loader2 className="h-5 w-5 animate-spin" /> : <Upload className="h-5 w-5" />}
            {uploading ? "جارٍ الرفع…" : "ارفع صورة إيصال التحويل"}
            <input type="file" accept="image/*" onChange={onFile} disabled={uploading} className="hidden" />
          </label>
        )}
        {error ? <p className="mt-2 text-sm text-rose-600">{error}</p> : null}
      </div>

      {waLink && (
        <a href={waLink} target="_blank" rel="noopener noreferrer" className="btn-primary mt-3 w-full">
          <MessageCircle className="h-5 w-5" /> تواصل معنا على واتساب
        </a>
      )}
      <p className="mt-3 text-center text-xs text-amber-700/80">
        بعد تأكيد دفعك سيظهر كود التفعيل هنا تلقائيًا. أبقِ الصفحة مفتوحة.
      </p>
    </div>
  );
}
