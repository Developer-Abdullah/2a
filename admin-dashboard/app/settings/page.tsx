import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";
import DashboardShell from "@/components/layout/DashboardShell";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export const dynamic = "force-dynamic";

export default async function SettingsPage() {
  const session = await getServerSession(authOptions);

  return (
    <DashboardShell title="title.settings">
      <div className="grid max-w-3xl gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Account</CardTitle>
            <CardDescription>The signed-in administrator.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-slate-500">Email</span>
              <span className="font-medium text-slate-900">{session?.user?.email ?? "—"}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-500">Role</span>
              <Badge variant="default">{session?.user?.role ?? "admin"}</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Tenant</CardTitle>
            <CardDescription>Data is isolated per tenant.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-slate-500">Active tenant</span>
              <span className="font-medium text-slate-900">store</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-500">Isolation</span>
              <span className="font-medium text-slate-900">dedicated_schema</span>
            </div>
            <p className="pt-2 text-sm text-slate-500">
              Store name, support email and branding are managed in the backend tenant configuration.
            </p>
          </CardContent>
        </Card>
      </div>
    </DashboardShell>
  );
}
