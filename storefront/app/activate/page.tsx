"use client";

import { useState } from "react";
import { KeyRound, Loader2, CheckCircle2, XCircle, Smartphone, Download, MessageCircle, ShieldCheck } from "lucide-react";
import type { CodeStatus } from "@/lib/types";
import { useStore } from "@/components/store-provider";

const WHATSAPP = (process.env.NEXT_PUBLIC_WHATSAPP || "").replace(/[^0-9]/g, "");

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
  const { t } = useStore();
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<CodeStatus | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [step, setStep] = useState<Step>("code");
  const [activating, setActivating] = useState(false);
  const [activateErr, setActivateErr] = useState("");
  const [app, setApp] = useState<AppInfo | null>(null);
  const [loadingApp, setLoadingApp] = useState(false);

  const deviceLabel = (d: string) => (d === "ipad" ? t("device.ipad") : d === "iphone" ? t("device.iphone") : t("device.both"));

  async function check(e: React.FormEvent) {
    e.preventDefault();
    if (!code.trim()) return;
    setLoading(true);
    setStatus(null);
    setNotFound(false);
    setStep("code");
    try {
      const res = await fetch(`/api/activation?code=${encodeURIComponent(code.trim())}`);
      if (res.status === 404) setNotFound(true);
      else if (res.ok) {
        const data = await res.json();
        setStatus(data.status);
        if (data.status?.valid) setStep("activate");
      } else setNotFound(true);
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
      } else if (res.status === 429) setActivateErr(t("activate.err_many"));
      else setActivateErr(t("activate.err_generic"));
    } catch {
      setActivateErr(t("activate.err_conn"));
    } finally {
      setActivating(false);
    }
  }

  async function loadApp() {
    setLoadingApp(true);
    try {
      const res = await fetch("/api/app");
      setApp(await res.json());
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
        <h1 className="mt-4 text-2xl font-extrabold text-foreground">{t("activate.title")}</h1>
        <p className="mt-2 text-muted-foreground">{t("activate.sub")}</p>
      </div>

      {/* Steps indicator */}
      <div className="mt-8 flex items-center justify-center gap-2 text-xs font-bold">
        <StepDot n={1} label={t("activate.step1")} active={step === "code"} done={step !== "code"} />
        <div className="h-px w-6 bg-border sm:w-8" />
        <StepDot n={2} label={t("activate.step2")} active={step === "activate"} done={step === "done"} />
        <div className="h-px w-6 bg-border sm:w-8" />
        <StepDot n={3} label={t("activate.step3")} active={step === "done"} done={false} />
      </div>

      {/* Step 1: enter code */}
      <form onSubmit={check} className="mt-8 flex gap-2">
        <input
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="XXXX-XXXX-XXXX"
          className="field flex-1 text-center font-mono text-lg tracking-widest"
        />
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? <Loader2 className="h-5 w-5 animate-spin" /> : <KeyRound className="h-5 w-5" />}
          {t("activate.check")}
        </button>
      </form>

      {notFound && (
        <div className="mt-6 flex items-center gap-3 rounded-2xl bg-rose-50 p-4 text-rose-700 dark:bg-rose-900/30 dark:text-rose-300">
          <XCircle className="h-6 w-6 flex-none" />
          <div>
            <p className="font-bold">{t("activate.notfound_t")}</p>
            <p className="text-sm">{t("activate.notfound_b")}</p>
          </div>
        </div>
      )}

      {status && !status.valid && (
        <div className="mt-6 flex items-center gap-3 rounded-2xl bg-amber-50 p-4 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300">
          <XCircle className="h-6 w-6 flex-none" />
          <p className="font-bold">
            {status.is_revoked ? t("activate.revoked") : status.expired ? t("activate.expired") : t("activate.used")}
          </p>
        </div>
      )}

      {/* Step 2: activate on this device */}
      {step === "activate" && status?.valid && (
        <div className="mt-6 card p-6">
          <div className="flex items-center gap-3 rounded-xl bg-emerald-50 p-3 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
            <CheckCircle2 className="h-6 w-6 flex-none" />
            <div>
              <p className="font-bold">{t("activate.valid")}</p>
              <p className="text-sm">{t("activate.devices")}: {status.current_device_count} / {status.max_devices} · {t("activate.type")}: {deviceLabel(status.device_type)}</p>
            </div>
          </div>
          <button onClick={activate} disabled={activating} className="btn-primary mt-5 w-full">
            {activating ? <Loader2 className="h-5 w-5 animate-spin" /> : <ShieldCheck className="h-5 w-5" />}
            {t("activate.activate_btn")}
          </button>
          {activateErr ? <p className="mt-2 text-center text-sm text-rose-600 dark:text-rose-400">{activateErr}</p> : null}
        </div>
      )}

      {/* Step 3: install */}
      {step === "done" && (
        <div className="mt-6 space-y-5">
          <div className="flex items-center gap-3 rounded-2xl bg-emerald-50 p-4 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
            <CheckCircle2 className="h-6 w-6 flex-none" />
            <p className="font-bold">{t("activate.done")}</p>
          </div>

          <div className="card p-6">
            <h2 className="flex items-center gap-2 text-lg font-extrabold text-foreground">
              <Smartphone className="h-5 w-5 text-brand-600 dark:text-brand-300" /> {t("activate.install_t")}
            </h2>

            {loadingApp ? (
              <div className="mt-4 flex items-center gap-2 text-muted-foreground"><Loader2 className="h-5 w-5 animate-spin" /> {t("activate.preparing")}</div>
            ) : app?.ready ? (
              <>
                <p className="mt-3 text-sm leading-7 text-muted-foreground">{t("activate.install_body")}</p>
                <a href={installHref} className="btn-primary mt-5 w-full">
                  <Download className="h-5 w-5" /> {t("activate.install_btn")} {app.name}{app.version ? ` (${app.version})` : ""}
                </a>
                <p className="mt-3 text-center text-xs text-muted-foreground">{t("activate.open_on_device")}</p>
              </>
            ) : (
              <div className="mt-4 rounded-xl bg-amber-50 p-4 text-sm text-amber-800 dark:bg-amber-900/20 dark:text-amber-300">
                {t("activate.not_ready")}
                {WHATSAPP && (
                  <a href={`https://wa.me/${WHATSAPP}`} target="_blank" rel="noopener noreferrer" className="btn-primary mt-3 w-full">
                    <MessageCircle className="h-5 w-5" /> {t("activate.whatsapp_help")}
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
          (done ? "bg-emerald-500" : active ? "bg-brand-600" : "bg-muted-foreground/50")
        }
      >
        {done ? <CheckCircle2 className="h-4 w-4" /> : n}
      </span>
      <span className={active ? "text-foreground" : "text-muted-foreground"}>{label}</span>
    </div>
  );
}
