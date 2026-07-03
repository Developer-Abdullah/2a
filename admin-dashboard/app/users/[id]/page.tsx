import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { UserDetail } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function UserDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  let detail: UserDetail | null = null;
  let error = "";
  try {
    detail = await adminGet<UserDetail>(`/v1/admin/users/${id}`);
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load user";
  }

  return (
    <DashboardShell title="title.user_details">
      <Link href="/users" className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowLeft className="mr-1 h-4 w-4" /> Back to users
      </Link>

      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : !detail ? (
        <p className="text-sm text-red-600">User not found.</p>
      ) : (
        <div className="space-y-6">
          <Card>
            <CardContent className="p-6">
              <h2 className="text-xl font-semibold text-slate-900">{detail.user.display_name || "Unknown user"}</h2>
              <p className="mt-1 font-mono text-xs text-slate-500">{detail.user.id}</p>
              <div className="mt-3 flex gap-6 text-sm text-slate-600">
                <span>{detail.user.device_count} device(s)</span>
                <span>Joined {detail.user.created_at.slice(0, 10)}</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Bound devices</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {detail.devices.length === 0 ? (
                <p className="px-6 pb-6 text-sm text-slate-500">No devices found.</p>
              ) : (
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
                    {detail.devices.map((d) => (
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
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
