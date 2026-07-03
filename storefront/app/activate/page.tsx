"use client";

import { useState } from "react";
import { KeyRound, Loader2, CheckCircle2, XCircle, Smartphone, Download, MessageCircle } from "lucide-react";
import type { CodeStatus } from "@/lib/types";

const WHATSAPP = (process.env.NEXT_PUBLIC_WHATSAPP || "").replace(/[^0-9]/g, "");

function deviceLabel(t: string) {
  return t === "ipad" ? "آيباد" : t === "iphone" ? "آيفون" : "آيفون / آيباد";
}

export default function ActivatePage() {
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<CodeStatus | null>(null);
  const [notFound, setNotFound] = useState(false);

  async function check(e: React.FormEvent) {
    e.preventDefault();
    if (!code.trim()) return;
    setLoading(true);
    setStatus(null);
    setNotFound(false);
    try {
      const res = await fetch(`/api/activation?code=${encodeURIComponent(code.trim())}`);
      if (res.status === 404) {
        setNotFound(true);
      } else if (res.ok) {
        const data = await res.json();
        setStatus(data.status);
      } else {
        setNotFound(true);
      }
    } catch {
      setNotFound(true);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="container-page max-w-2xl py-12">
      <div className="text-center">
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl brand-gradient text-white">
          <KeyRound className="h-7 w-7" />
        </div>
        <h1 className="mt-4 text-2xl font-extrabold text-ink">تفعيل الاشتراك</h1>
        <p className="mt-2 text-slate-500">أدخل كود التفعيل الذي وصلك للتحقق من حالته ومعرفة خطوات التثبيت.</p>
      </div>

      <form onSubmit={check} className="mt-8 flex gap-2">
        <input
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="XXXX-XXXX-XXXX"
          className="flex-1 rounded-xl border border-slate-200 bg-white px-4 py-3 text-center font-mono text-lg tracking-widest outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
        />
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? <Loader2 className="h-5 w-5 animate-spin" /> : <KeyRound className="h-5 w-5" />}
          تحقّق
        </button>
      </form>

      {notFound && (
        <div className="mt-6 flex items-center gap-3 rounded-2xl bg-rose-50 p-4 text-rose-700">
          <XCircle className="h-6 w-6 flex-none" />
          <div>
            <p className="font-bold">الكود غير موجود</p>
            <p className="text-sm">تأكد من كتابة الكود بشكل صحيح كما وصلك تمامًا.</p>
          </div>
        </div>
      )}

      {status && (
        <div className="mt-6 space-y-5">
          <div
            className={
              "flex items-center gap-3 rounded-2xl p-4 " +
              (status.valid ? "bg-emerald-50 text-emerald-700" : "bg-amber-50 text-amber-800")
            }
          >
            {status.valid ? <CheckCircle2 className="h-6 w-6 flex-none" /> : <XCircle className="h-6 w-6 flex-none" />}
            <div>
              <p className="font-bold">
                {status.valid
                  ? "الكود صالح وجاهز للتفعيل ✅"
                  : status.is_revoked
                  ? "هذا الكود موقوف"
                  : status.expired
                  ? "انتهت صلاحية هذا الكود"
                  : "تم استخدام هذا الكود بالكامل"}
              </p>
              <p className="text-sm">
                الأجهزة: {status.current_device_count} / {status.max_devices} · النوع: {deviceLabel(status.device_type)}
              </p>
            </div>
          </div>

          {status.valid && (
            <div className="card p-6">
              <h2 className="flex items-center gap-2 text-lg font-extrabold text-ink">
                <Smartphone className="h-5 w-5 text-brand-600" /> خطوات التفعيل على جهازك
              </h2>
              <ol className="mt-4 space-y-3 text-sm leading-7 text-slate-600">
                <li className="flex gap-3">
                  <span className="flex h-6 w-6 flex-none items-center justify-center rounded-full bg-brand-100 text-xs font-bold text-brand-700">1</span>
                  افتح صفحة التثبيت من جهاز {deviceLabel(status.device_type)} الذي تريد تفعيله.
                </li>
                <li className="flex gap-3">
                  <span className="flex h-6 w-6 flex-none items-center justify-center rounded-full bg-brand-100 text-xs font-bold text-brand-700">2</span>
                  أدخل كود التفعيل عند الطلب، وستُثبَّت شهادة الوصول على جهازك.
                </li>
                <li className="flex gap-3">
                  <span className="flex h-6 w-6 flex-none items-center justify-center rounded-full bg-brand-100 text-xs font-bold text-brand-700">3</span>
                  بعد التثبيت تظهر مكتبة التطبيقات، ثبّت ما تريد واستمتع بالاشتراك طوال مدته.
                </li>
              </ol>
              {WHATSAPP && (
                <a
                  href={`https://wa.me/${WHATSAPP}?text=${encodeURIComponent("مرحبًا، أحتاج رابط التثبيت لتفعيل اشتراكي.")}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn-primary mt-5 w-full"
                >
                  <Download className="h-5 w-5" /> احصل على رابط التثبيت عبر واتساب
                </a>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
