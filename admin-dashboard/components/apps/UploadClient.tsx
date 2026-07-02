"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { getUploadUrlAction, createVersionAction } from "@/lib/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

// Upload & Sign: requests a presigned PUT URL, uploads the raw IPA to S3 (best effort — falls
// back gracefully if object storage is unreachable), then enqueues a signing job for the version.
export default function UploadClient({ appId }: { appId: string }) {
  const router = useRouter();
  const [version, setVersion] = useState("");
  const [build, setBuild] = useState("1");
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [pending, setPending] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setErr("");
    setMsg("");
    try {
      let rawKey = `manual/${appId}/${version}.ipa`;
      let size = 0;
      let note = "";
      if (file) {
        size = file.size;
        try {
          const { presigned_url, expected_key } = await getUploadUrlAction(appId);
          rawKey = expected_key;
          const put = await fetch(presigned_url, { method: "PUT", body: file });
          note = put.ok ? "Uploaded to storage. " : "Upload skipped (storage unreachable). ";
        } catch {
          note = "Upload skipped (storage unreachable). ";
        }
      }
      const fd = new FormData();
      fd.set("app_id", appId);
      fd.set("version", version);
      fd.set("build_number", build);
      fd.set("raw_ipa_s3_key", rawKey);
      fd.set("size_bytes", String(size));
      const res = await createVersionAction({}, fd);
      if (res.ok) {
        setMsg(note + "Version queued for signing.");
        setVersion("");
        setFile(null);
        router.refresh();
      } else {
        setErr(res.error || "Failed to create version");
      }
    } catch (e2) {
      setErr(e2 instanceof Error ? e2.message : "Failed");
    }
    setPending(false);
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1">
          <label className="text-sm font-medium text-slate-700">Version</label>
          <Input value={version} onChange={(e) => setVersion(e.target.value)} required placeholder="1.0.0" />
        </div>
        <div className="space-y-1">
          <label className="text-sm font-medium text-slate-700">Build</label>
          <Input value={build} onChange={(e) => setBuild(e.target.value)} placeholder="1" />
        </div>
      </div>
      <div className="space-y-1">
        <label className="text-sm font-medium text-slate-700">IPA file (optional)</label>
        <Input type="file" accept=".ipa" onChange={(e) => setFile(e.target.files?.[0] ?? null)} />
      </div>
      {err ? <p className="text-sm text-red-600">{err}</p> : null}
      {msg ? <p className="text-sm text-green-600">{msg}</p> : null}
      <Button type="submit" disabled={pending}>{pending ? "Submitting…" : "Upload & Sign"}</Button>
    </form>
  );
}
