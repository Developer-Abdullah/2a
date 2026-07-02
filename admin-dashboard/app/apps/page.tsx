import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AppItem } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

export default async function AppsPage() {
  let apps: AppItem[] = [];
  let error = "";
  try {
    const data = await adminGet<{ items: AppItem[] }>("/v1/admin/apps");
    apps = data.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load applications";
  }

  return (
    <DashboardShell title="Applications">
      <div className="mb-6 flex items-center justify-between">
        <p className="text-sm text-slate-500">{apps.length} application(s)</p>
        <Link href="/apps/new">
          <Button>New App</Button>
        </Link>
      </div>

      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : apps.length === 0 ? (
        <p className="text-sm text-slate-500">No apps yet. Create your first one.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Bundle ID</TableHead>
                  <TableHead>Category</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {apps.map((a) => (
                  <TableRow key={a.id}>
                    <TableCell className="font-medium text-slate-900">{a.name}</TableCell>
                    <TableCell className="text-slate-500">{a.bundle_identifier}</TableCell>
                    <TableCell className="text-slate-600">{a.category}</TableCell>
                    <TableCell>
                      {a.is_published ? <Badge variant="success">Published</Badge> : <Badge variant="muted">Draft</Badge>}
                    </TableCell>
                    <TableCell className="text-right">
                      <Link href={`/apps/${a.id}`} className="text-sm font-medium text-blue-600 hover:underline">
                        Manage
                      </Link>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </DashboardShell>
  );
}
