"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createTenantAction, type ActionState } from "@/lib/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export default function CreateMerchantForm() {
  const router = useRouter();
  const [state, setState] = useState<ActionState>({});
  const [pending, setPending] = useState(false);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    setPending(true);
    const res = await createTenantAction({}, fd);
    setPending(false);
    setState(res);
    if (res.ok) {
      router.push("/owner");
      router.refresh();
    }
  }

  return (
    <form onSubmit={onSubmit} className="max-w-xl space-y-4">
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Store name</label>
        <Input name="name" required placeholder="Azzam Store" />
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Slug</label>
        <Input name="slug" required placeholder="azzam" pattern="[a-z0-9-]+" />
        <p className="text-xs text-slate-500">Lowercase letters, digits and hyphens. Becomes the tenant&apos;s schema.</p>
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Merchant admin email</label>
        <Input name="admin_email" type="email" required placeholder="owner@azzam.com" />
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Admin password</label>
        <Input name="admin_password" type="password" required minLength={6} placeholder="At least 6 characters" />
      </div>
      {state.error ? <p className="text-sm text-red-600">{state.error}</p> : null}
      {state.message ? <p className="text-sm text-green-600">{state.message}</p> : null}
      <Button type="submit" disabled={pending}>{pending ? "Provisioning…" : "Create merchant"}</Button>
    </form>
  );
}
