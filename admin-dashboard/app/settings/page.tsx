import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";
import DashboardShell from "@/components/layout/DashboardShell";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

export default async function SettingsPage() {
  const t = await getT();
  const session = await getServerSession(authOptions);

  return (
    <DashboardShell title="title.settings">
      <div className="grid max-w-3xl gap-6">
        <Card>
          <CardHeader>
            <CardTitle>{t("set.account")}</CardTitle>
            <CardDescription>{t("set.account_desc")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">{t("set.email")}</span>
              <span className="font-medium text-foreground">{session?.user?.email ?? "—"}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">{t("set.role")}</span>
              <Badge variant="default">{session?.user?.role ?? "admin"}</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("set.tenant")}</CardTitle>
            <CardDescription>{t("set.tenant_desc")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">{t("set.active_tenant")}</span>
              <span className="font-medium text-foreground">store</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">{t("set.isolation")}</span>
              <span className="font-medium text-foreground">dedicated_schema</span>
            </div>
            <p className="pt-2 text-sm text-muted-foreground">{t("set.tenant_note")}</p>
          </CardContent>
        </Card>
      </div>
    </DashboardShell>
  );
}
