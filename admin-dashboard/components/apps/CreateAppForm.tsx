"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createAppAction, type ActionState } from "@/lib/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const CATEGORIES = [
  "exclusive", "social_media", "modified_apps", "modified_games", "paid_apps",
  "paid_games", "arcade_games", "design", "productivity", "entertainment",
  "sports", "islamic", "jailbreak", "in_house", "other",
];

export default function CreateAppForm() {
  const router = useRouter();
  const [state, setState] = useState<ActionState>({});
  const [pending, setPending] = useState(false);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = e.currentTarget;
    const fd = new FormData(form);
    setPending(true);
    const res = await createAppAction({}, fd);
    setPending(false);
    setState(res);
    if (res.ok && res.data?.id) {
      router.push(`/apps/${res.data.id}`);
    }
  }

  return (
    <form onSubmit={onSubmit} className="max-w-xl space-y-4">
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Name</label>
        <Input name="name" required placeholder="Cloud Notes" />
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Bundle Identifier</label>
        <Input name="bundle_identifier" required placeholder="com.company.app" />
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Category</label>
        <select name="category" defaultValue="other" className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm">
          {CATEGORIES.map((c) => <option key={c} value={c}>{c}</option>)}
        </select>
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Description</label>
        <textarea name="description" rows={3} className="flex w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm" placeholder="What does this app do?" />
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">Features (comma separated)</label>
        <Input name="features" placeholder="Fast sync, Offline mode" />
      </div>
      <label className="flex items-center gap-2 text-sm text-slate-700">
        <input type="checkbox" name="publish" defaultChecked /> Publish immediately
      </label>
      {state.error ? <p className="text-sm text-red-600">{state.error}</p> : null}
      {state.message ? <p className="text-sm text-green-600">{state.message}</p> : null}
      <Button type="submit" disabled={pending}>{pending ? "Creating…" : "Create app"}</Button>
    </form>
  );
}
