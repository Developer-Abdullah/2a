import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { ActivationCode } from "@/lib/types";
import GenerateCodesForm from "@/components/activation/GenerateCodesForm";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function ActivationPage() {
  let items: ActivationCode[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: ActivationCode[] }>("/v1/admin/activation/codes");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load codes";
  }

  return (
    <DashboardShell title="title.activation_codes">
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Generate codes</CardTitle>
          </CardHeader>
          <CardContent>
            <GenerateCodesForm />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Codes ({items.length})</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {error ? (
              <p className="px-6 pb-6 text-sm text-red-600">{error}</p>
            ) : items.length === 0 ? (
              <p className="px-6 pb-6 text-sm text-slate-500">No codes yet.</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Code</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Device</TableHead>
                    <TableHead>Uses</TableHead>
                    <TableHead>Devices</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="font-mono text-sm">{c.code}</TableCell>
                      <TableCell className="text-slate-600">{c.type}</TableCell>
                      <TableCell className="text-slate-600">{c.device_type}</TableCell>
                      <TableCell className="text-slate-500">{c.current_uses}/{c.max_uses}</TableCell>
                      <TableCell className="text-slate-500">{c.current_device_count}/{c.max_devices}</TableCell>
                      <TableCell>
                        {c.is_revoked ? <Badge variant="danger">Revoked</Badge> : <Badge variant="success">Active</Badge>}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </DashboardShell>
  );
}
