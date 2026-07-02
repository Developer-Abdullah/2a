import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import CreateAppForm from "@/components/apps/CreateAppForm";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export const dynamic = "force-dynamic";

export default function NewAppPage() {
  return (
    <DashboardShell title="New Application">
      <Link href="/apps" className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowLeft className="mr-1 h-4 w-4" /> Back to apps
      </Link>
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Create application</CardTitle>
          <CardDescription>Register a new app in the store.</CardDescription>
        </CardHeader>
        <CardContent>
          <CreateAppForm />
        </CardContent>
      </Card>
    </DashboardShell>
  );
}
