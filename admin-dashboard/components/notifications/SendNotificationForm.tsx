"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { sendNotificationAction, type ActionState } from "@/lib/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useI18n } from "@/components/providers/LocaleProvider";

export default function SendNotificationForm() {
  const { t } = useI18n();
  const router = useRouter();
  const [state, setState] = useState<ActionState>({});
  const [pending, setPending] = useState(false);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = e.currentTarget;
    const fd = new FormData(form);
    setPending(true);
    const res = await sendNotificationAction({}, fd);
    setPending(false);
    setState(res);
    if (res.ok) {
      form.reset();
      router.refresh();
    }
  }

  return (
    <form onSubmit={onSubmit} className="max-w-xl space-y-4">
      <div className="space-y-1">
        <label className="text-sm font-medium text-foreground/80">{t("send.title")}</label>
        <Input name="title" required placeholder="New release available" />
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-foreground/80">{t("send.body")}</label>
        <textarea name="body" rows={3} required className="flex w-full rounded-md border border-input bg-card px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" placeholder="Tap to update to the latest version." />
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-foreground/80">{t("send.type")}</label>
        <select name="type" defaultValue="alert" className="flex h-10 w-full rounded-md border border-input bg-card px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <option value="alert">alert</option>
          <option value="silent">silent</option>
          <option value="update">update</option>
        </select>
      </div>
      {state.error ? <p className="text-sm text-red-600">{state.error}</p> : null}
      {state.message ? <p className="text-sm text-green-600">{state.message}</p> : null}
      <Button type="submit" disabled={pending}>{pending ? t("send.sending") : t("send.send")}</Button>
    </form>
  );
}
