import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { ActivationCode } from "@/lib/types";
import GenerateCodesForm from "@/components/activation/GenerateCodesForm";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

export default async function ActivationPage() {
  const t = await getT();
  let items: ActivationCode[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: ActivationCode[] }>("/v1/admin/activation/codes");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  return (
    <DashboardShell title="title.activation_codes">
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{t("act.generate")}</CardTitle>
          </CardHeader>
          <CardContent>
            <GenerateCodesForm />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("act.codes")} ({items.length})</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {error ? (
              <p className="px-6 pb-6 text-sm text-red-600 dark:text-red-400">{error}</p>
            ) : items.length === 0 ? (
              <p className="px-6 pb-6 text-sm text-muted-foreground">{t("act.no_codes")}</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("act.col.code")}</TableHead>
                    <TableHead>{t("act.col.type")}</TableHead>
                    <TableHead>{t("act.col.device")}</TableHead>
                    <TableHead>{t("act.col.uses")}</TableHead>
                    <TableHead>{t("act.col.devices")}</TableHead>
                    <TableHead>{t("common.status")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="font-mono text-sm text-foreground">{c.code}</TableCell>
                      <TableCell className="text-foreground/80">{c.type}</TableCell>
                      <TableCell className="text-foreground/80">{c.device_type}</TableCell>
                      <TableCell className="text-muted-foreground">{c.current_uses}/{c.max_uses}</TableCell>
                      <TableCell className="text-muted-foreground">{c.current_device_count}/{c.max_devices}</TableCell>
                      <TableCell>
                        {c.is_revoked ? <Badge variant="danger">{t("common.revoked")}</Badge> : <Badge variant="success">{t("common.active")}</Badge>}
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
