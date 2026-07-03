import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { AdminDevice } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

export default async function EnrollmentPage() {
  const t = await getT();
  let items: AdminDevice[] = [];
  let error = "";
  try {
    const d = await adminGet<{ items: AdminDevice[] }>("/v1/admin/devices");
    items = d.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  const fingerprint = items.filter((d) => d.enrollment_method === "fingerprint").length;
  const udid = items.filter((d) => d.enrollment_method === "udid_profile").length;

  return (
    <DashboardShell title="title.enrollment">
      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : (
        <div className="space-y-6">
          <div className="grid grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{t("enr.fingerprint")}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-foreground">{fingerprint}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{t("enr.udid")}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-foreground">{udid}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{t("enr.total")}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-foreground">{items.length}</p>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>{t("enr.how_title")}</CardTitle>
              <CardDescription>{t("enr.how_desc")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-muted-foreground">
              <p>
                1. {t("enr.step1a")}
                <Link href="/activation" className="text-brand-700 hover:underline dark:text-brand-300">
                  {t("enr.step1_link")}
                </Link>
                .
              </p>
              <p>2. {t("enr.step2")}</p>
              <p>3. {t("enr.step3")}</p>
            </CardContent>
          </Card>

          {items.length > 0 ? (
            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("common.device")}</TableHead>
                      <TableHead>{t("common.model")}</TableHead>
                      <TableHead>{t("enr.col.method")}</TableHead>
                      <TableHead>{t("common.status")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((d) => (
                      <TableRow key={d.id}>
                        <TableCell className="font-medium capitalize text-foreground">{d.device_type}</TableCell>
                        <TableCell className="text-muted-foreground">{d.model || "—"}</TableCell>
                        <TableCell>
                          <Badge variant="muted">{d.enrollment_method}</Badge>
                        </TableCell>
                        <TableCell>
                          {d.is_revoked ? <Badge variant="danger">{t("common.revoked")}</Badge> : <Badge variant="success">{t("common.active")}</Badge>}
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
