"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { generateCodesAction, type ActionState } from "@/lib/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const selectClass = "flex h-10 w-full rounded-md border border-input bg-card px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring";

export default function GenerateCodesForm() {
  const router = useRouter();
  const [state, setState] = useState<ActionState>({});
  const [pending, setPending] = useState(false);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    setPending(true);
    const res = await generateCodesAction({}, fd);
    setPending(false);
    setState(res);
    if (res.ok) router.refresh();
  }

  const codes = (state.data?.codes as string[] | undefined) ?? [];

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1">
          <label className="text-sm font-medium text-foreground/80">Count</label>
          <Input name="count" type="number" min={1} max={500} defaultValue={5} />
        </div>
        <div className="space-y-1">
          <label className="text-sm font-medium text-foreground/80">Code type</label>
          <select name="code_type" defaultValue="usage_count" className={selectClass}>
            <option value="usage_count">usage_count</option>
            <option value="time_bound">time_bound</option>
            <option value="device_bound">device_bound</option>
            <option value="reseller_bulk">reseller_bulk</option>
          </select>
        </div>
        <div className="space-y-1">
          <label className="text-sm font-medium text-foreground/80">Device type</label>
          <select name="device_type" defaultValue="both" className={selectClass}>
            <option value="both">both</option>
            <option value="iphone">iphone</option>
            <option value="ipad">ipad</option>
          </select>
        </div>
        <div className="space-y-1">
          <label className="text-sm font-medium text-foreground/80">Max devices</label>
          <Input name="max_devices" type="number" min={1} defaultValue={1} />
        </div>
        <div className="space-y-1">
          <label className="text-sm font-medium text-foreground/80">Max uses</label>
          <Input name="max_uses" type="number" min={1} defaultValue={1} />
        </div>
      </div>
      {state.error ? <p className="text-sm text-red-600">{state.error}</p> : null}
      {state.message ? <p className="text-sm text-green-600">{state.message}</p> : null}
      {codes.length > 0 ? (
        <div className="rounded-md border border-border bg-muted p-3">
          <p className="mb-2 text-xs font-medium uppercase text-muted-foreground">Generated codes</p>
          <div className="flex flex-wrap gap-2">
            {codes.map((c) => (
              <code key={c} className="rounded bg-card px-2 py-1 text-sm text-foreground shadow-sm">{c}</code>
            ))}
          </div>
        </div>
      ) : null}
      <Button type="submit" disabled={pending}>{pending ? "Generating…" : "Generate codes"}</Button>
    </form>
  );
}
