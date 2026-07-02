import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { ActivationCode } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export const dynamic = "force-dynamic";

export default async function ActivationDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  let code: ActivationCode | undefined;
  let error = "";
  try {
    const d = await adminGet<{ items: ActivationCode[] }>("/v1/admin/activation/codes");
    code = d.items.find((c) => c.id === id);
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load code";
  }

  return (
    <DashboardShell title="Activation Code">
      <Link href="/activation" className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowLeft className="mr-1 h-4 w-4" /> Back to codes
      </Link>

      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : !code ? (
        <p className="text-sm text-red-600">Code not found.</p>
      ) : (
        <Card className="max-w-2xl">
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <CardTitle className="font-mono">{code.code}</CardTitle>
            {code.is_revoked ? <Badge variant="danger">Revoked</Badge> : <Badge variant="success">Active</Badge>}
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-slate-500">Type:</span>{" "}
                <span className="font-medium text-slate-900">{code.type}</span>
              </div>
              <div>
                <span className="text-slate-500">Allowed devices:</span>{" "}
                <span className="font-medium capitalize text-slate-900">{code.device_type}</span>
              </div>
              <div>
                <span className="text-slate-500">Devices:</span>{" "}
                <span className="font-medium text-slate-900">{code.current_device_count} / {code.max_devices}</span>
              </div>
              <div>
                <span className="text-slate-500">Uses:</span>{" "}
                <span className="font-medium text-slate-900">{code.current_uses} / {code.max_uses}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </DashboardShell>
  );
}
