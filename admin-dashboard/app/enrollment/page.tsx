import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminDevice } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function EnrollmentPage() {
  let items: AdminDevice[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AdminDevice[] }>("/v1/admin/devices");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load enrollment data";
  }

  const fingerprint = items.filter((d) => d.enrollment_method === "fingerprint").length;
  const udid = items.filter((d) => d.enrollment_method === "udid_profile").length;

  return (
    <DashboardShell title="Enrollment">
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
        <div className="space-y-6">
          <div className="grid grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-slate-500">Fingerprint</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-slate-900">{fingerprint}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-slate-500">UDID profile</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-slate-900">{udid}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-slate-500">Total</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-slate-900">{items.length}</p>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>How enrollment works</CardTitle>
              <CardDescription>From activation code to a validated device.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-slate-600">
              <p>
                1. A user redeems an{" "}
                <Link href="/activation" className="text-blue-600 hover:underline">
                  activation code
                </Link>
                .
              </p>
              <p>2. The device is enrolled by fingerprint hash or by installing a UDID provisioning profile.</p>
              <p>3. The backend validates the device and binds it to the user&apos;s code.</p>
            </CardContent>
          </Card>

          {items.length > 0 ? (
            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Device</TableHead>
                      <TableHead>Model</TableHead>
                      <TableHead>Method</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((d) => (
                      <TableRow key={d.id}>
                        <TableCell className="font-medium capitalize text-slate-900">{d.device_type}</TableCell>
                        <TableCell className="text-slate-500">{d.model || "—"}</TableCell>
                        <TableCell>
                          <Badge variant="muted">{d.enrollment_method}</Badge>
                        </TableCell>
                        <TableCell>
                          {d.is_revoked ? <Badge variant="danger">Revoked</Badge> : <Badge variant="success">Active</Badge>}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          ) : null}
        </div>
      )}
    </DashboardShell>
  );
}
