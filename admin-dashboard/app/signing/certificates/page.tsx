import { ShieldCheck } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { Certificate } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function CertificatesPage() {
  let items: Certificate[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: Certificate[] }>("/v1/admin/certificates");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load certificates";
  }
  const active = items.filter((c) => c.is_active).length;

  return (
    <DashboardShell title="Certificate Pool">
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
        <div className="space-y-6">
          <Card>
            <CardHeader className="flex-row items-center gap-3 space-y-0">
              <ShieldCheck className="h-6 w-6 text-emerald-600" />
              <div>
                <CardTitle className="text-base">Pool health</CardTitle>
                <CardDescription>{active} active certificate(s). Apple Enterprise signing certificates.</CardDescription>
              </div>
            </CardHeader>
          </Card>

          <Card>
            <CardContent className="p-0">
              {items.length === 0 ? (
                <p className="p-6 text-sm text-slate-500">No certificates uploaded.</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Label</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Added</TableHead>
                      <TableHead>Expires</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((c) => (
                      <TableRow key={c.id}>
                        <TableCell className="font-medium text-slate-900">{c.label}</TableCell>
                        <TableCell>
                          {c.is_active ? <Badge variant="success">Active</Badge> : <Badge variant="muted">Inactive</Badge>}
                        </TableCell>
                        <TableCell className="text-slate-500">{c.added_at.slice(0, 10)}</TableCell>
                        <TableCell className="text-slate-500">{c.expires_at.slice(0, 10)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
