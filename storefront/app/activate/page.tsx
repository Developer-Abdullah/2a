"use client";

import { useState } from "react";
import { KeyRound, Loader2, CheckCircle2, XCircle, Smartphone, Download, MessageCircle, ShieldCheck } from "lucide-react";
import type { CodeStatus } from "@/lib/types";

const WHATSAPP = (process.env.NEXT_PUBLIC_WHATSAPP || "").replace(/[^0-9]/g, "");

function deviceLabel(t: string) {
  return t === "ipad" ? "آيباد" : t === "iphone" ? "آيفون" : "آيفون / آيباد";
}

// A stable per-browser device id so re-activations from the same device don't consume extra slots.
function getDeviceId(): string {
  if (typeof window === "undefined") return "web";
  let id = localStorage.getItem("da_device_id");
  if (!id) {
    id = (crypto as Crypto & { randomUUID?: () => string }).randomUUID?.() || `web-${Date.now()}-${Math.random().toString(36).slice(2)}`;
    localStorage.setItem("da_device_id", id);
  }
  return id;
}

function guessDeviceType(): string {
  if (typeof navigator === "undefined") return "iphone";
  return /ipad/i.test(navigator.userAgent) ? "ipad" : "iphone";
}

interface AppInfo {
  ready: boolean;
  name?: string;
  version?: string;
  manifest_url?: string;
}

type Step = "code" | "activate" | "done";

export default function ActivatePage() {
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<CodeStatus | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [step, setStep] = useState<Step>("code");
  const [activating, setActivating] = useState(false);
  const [activateErr, setActivateErr] = useState("");
  const [app, setApp] = useState<AppInfo | null>(null);
  const [loadingApp, setLoadingApp] = useState(false);

  async function check(e: React.FormEvent) {
    e.preventDefault();
    if (!code.trim()) return;
    setLoading(true);
    setStatus(null);
    setNotFound(false);
    setStep("code");
    try {
      const res = await fetch(`/api/activation?code=${encodeURIComponent(code.trim())}`);
      if (res.status === 404) {
        setNotFound(true);
      } else if (res.ok) {
        const data = await res.json();
        setStatus(data.status);
        if (data.status?.valid) setStep("activate");
      } else {
        setNotFound(true);
      }
    } catch {
      setNotFound(true);
    } finally {
      setLoading(false);
    }
  }

  async function activate() {
    setActivateErr("");
    setActivating(true);
    try {
      const res = await fetch("/api/activate-device", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code: code.trim(), device_id: getDeviceId(), device_type: guessDeviceType() }),
      });
      if (res.ok) {
        setStep("done");
        loadApp();
      } else if (res.status === 429) {
        setActivateErr("محاولات كثيرة. انتظر دقيقة وحاول مجددًا.");
      } else {
        setActivateErr("تعذّر تفعيل الكود على هذا الجهاز. تأكد أنه صالح ولم تُستهلك حصته.");
      }
    } catch {
      setActivateErr("تعذّر الاتصال. حاول مرة أخرى.");
    } finally {
      setActivating(false);
    }
  }

  async function loadApp() {
    setLoadingApp(true);
    try {
      const res = await fetch("/api/app");
      const data = await res.json();
      setApp(data);
    } catch {
      setApp({ ready: false });
    } finally {
      setLoadingApp(false);
    }
  }

  const installHref = app?.manifest_url
    ? `itms-services://?action=download-manifest&url=${encodeURIComponent(app.manifest_url)}`
    : "";

  return (
    <div className="container-page max-w-2xl py-12">
      <div className="text-center">
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl brand-gradient text-white">
          <KeyRound className="h-7 w-7" />
        </div>
        <h1 className="mt-4 text-2xl font-extrabold text-ink">تفعيل الاشتراك</h1>
        <p className="mt-2 text-slate-500">أدخل كود التفعيل، فعّله على جهازك، ثم ثبّت التطبيق.</p>
      </div>

      {/* Steps indicator */}
      <div className="mt-8 flex items-center justify-center gap-2 text-xs font-bold">
        <StepDot n={1} label="الكود" active={step === "code"} done={step !== "code"} />
        <div className="h-px w-8 bg-slate-200" />
        <StepDot n={2} label="التفعيل" active={step === "activate"} done={step === "done"} />
        <div className="h-px w-8 bg-slate-200" />
        <StepDot n={3} label="التثبيت" active={step === "done"} done={false} />
      </div>

      {/* Step 1: enter code */}
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
            <p className="text-sm">تأكد من كتابة الكود كما وصلك تمامًا.</p>
          </div>
        </div>
      )}

      {status && !status.valid && (
        <div className="mt-6 flex items-center gap-3 rounded-2xl bg-amber-50 p-4 text-amber-800">
          <XCircle className="h-6 w-6 flex-none" />
          <p className="font-bold">
            {status.is_revoked ? "هذا الكود موقوف" : status.expired ? "انتهت صلاحية هذا الكود" : "تم استخدام هذا الكود بالكامل"}
          </p>
        </div>
      )}

      {/* Step 2: activate on this device */}
      {step === "activate" && status?.valid && (
        <div className="mt-6 card p-6">
          <div className="flex items-center gap-3 rounded-xl bg-emerald-50 p-3 text-emerald-700">
            <CheckCircle2 className="h-6 w-6 flex-none" />
            <div>
              <p className="font-bold">الكود صالح ✅</p>
              <p className="text-sm">الأجهزة: {status.current_device_count} / {status.max_devices} · النوع: {deviceLabel(status.device_type)}</p>
            </div>
          </div>
          <button onClick={activate} disabled={activating} className="btn-primary mt-5 w-full">
            {activating ? <Loader2 className="h-5 w-5 animate-spin" /> : <ShieldCheck className="h-5 w-5" />}
            فعّل على هذا الجهاز
          </button>
          {activateErr ? <p className="mt-2 text-center text-sm text-rose-600">{activateErr}</p> : null}
        </div>
      )}

      {/* Step 3: install */}
      {step === "done" && (
        <div className="mt-6 space-y-5">
          <div className="flex items-center gap-3 rounded-2xl bg-emerald-50 p-4 text-emerald-700">
            <CheckCircle2 className="h-6 w-6 flex-none" />
            <p className="font-bold">تم تفعيل الاشتراك على جهازك 🎉</p>
          </div>

          <div className="card p-6">
            <h2 className="flex items-center gap-2 text-lg font-extrabold text-ink">
              <Smartphone className="h-5 w-5 text-brand-600" /> ثبّت التطبيق
            </h2>

            {loadingApp ? (
              <div className="mt-4 flex items-center gap-2 text-slate-500"><Loader2 className="h-5 w-5 animate-spin" /> جارٍ التحضير…</div>
            ) : app?.ready ? (
              <>
                <p className="mt-3 text-sm leading-7 text-slate-600">
                  اضغط زر التثبيت من جهاز {deviceLabel(status?.device_type || "iphone")}، ووافق على تثبيت التطبيق. بعد التثبيت
                  ستجد بداخله مكتبة التطبيقات جاهزة للاستخدام طوال مدة اشتراكك.
                </p>
                <a href={installHref} className="btn-primary mt-5 w-full">
                  <Download className="h-5 w-5" /> تثبيت {app.name}{app.version ? ` (${app.version})` : ""}
                </a>
                <p className="mt-3 text-center text-xs text-slate-400">
                  افتح هذه الصفحة من جهاز الآيفون/الآيباد نفسه ليعمل زر التثبيت.
                </p>
              </>
            ) : (
              <div className="mt-4 rounded-xl bg-amber-50 p-4 text-sm text-amber-800">
                التطبيق قيد التجهيز حاليًا. تواصل معنا وسنرسل لك رابط التثبيت مباشرة.
                {WHATSAPP && (
                  <a
                    href={`https://wa.me/${WHATSAPP}?text=${encodeURIComponent("مرحبًا، فعّلت اشتراكي وأحتاج رابط التثبيت.")}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="btn-primary mt-3 w-full"
                  >
                    <MessageCircle className="h-5 w-5" /> تواصل عبر واتساب
                  </a>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function StepDot({ n, label, active, done }: { n: number; label: string; active: boolean; done: boolean }) {
  return (
    <div className="flex items-center gap-1.5">
      <span
        className={
          "flex h-6 w-6 items-center justify-center rounded-full text-white " +
          (done ? "bg-emerald-500" : active ? "bg-brand-600" : "bg-slate-300")
        }
      >
        {done ? <CheckCircle2 className="h-4 w-4" /> : n}
      </span>
      <span className={active ? "text-ink" : "text-slate-400"}>{label}</span>
    </div>
  );
}
