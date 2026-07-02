import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AppItem, AppVersion } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge, type BadgeVariant } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

function statusVariant(s: string): BadgeVariant {
  if (s === "signed") return "success";
  if (s === "failed") return "danger";
  return "warning";
}

export default async function AppDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  let app: AppItem | undefined;
  let versions: AppVersion[] = [];
  let error = "";
  try {
    const apps = await adminGet<{ items: AppItem[] }>("/v1/admin/apps");
    app = apps.items.find((a) => a.id === id);
    const v = await adminGet<{ items: AppVersion[] }>("/v1/admin/apps/" + id + "/versions");
    versions = v.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load app";
  }

  return (
    <DashboardShell title="App Details">
      <Link href="/apps" className="mb-6 inline-block text-sm text-slate-500 hover:text-slate-900">
        ← Back to apps
      </Link>

      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : !app ? (
        <p className="text-sm text-red-600">App not found.</p>
      ) : (
        <div className="space-y-6">
          <Card>
            <CardContent className="flex items-center justify-between p-6">
              <div>
                <h2 className="text-xl font-semibold text-slate-900">{app.name}</h2>
                <p className="text-sm text-slate-500">{app.bundle_identifier}</p>
                <div className="mt-2 flex items-center gap-2">
                  <Badge variant="muted">{app.category}</Badge>
                  {app.is_published ? <Badge variant="success">Published</Badge> : <Badge variant="muted">Draft</Badge>}
                </div>
              </div>
              <Link href={`/apps/${id}/versions/new`}>
                <Button>Upload &amp; Sign</Button>
              </Link>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Versions</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {versions.length === 0 ? (
                <p className="px-6 pb-6 text-sm text-slate-500">No versions yet. Upload one to start signing.</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Version</TableHead>
                      <TableHead>Build</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Size</TableHead>
                      <TableHead>Created</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {versions.map((v) => (
                      <TableRow key={v.id}>
                        <TableCell className="font-medium text-slate-900">{v.version}</TableCell>
                        <TableCell className="text-slate-500">{v.build_number}</TableCell>
                        <TableCell>
                          <Badge variant={statusVariant(v.signing_status)}>{v.signing_status}</Badge>
                        </TableCell>
                        <TableCell className="text-slate-500">{(v.size_bytes / 1_000_000).toFixed(1)} MB</TableCell>
                        <TableCell className="text-slate-500">{v.created_at.slice(0, 10)}</TableCell>
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
