import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import UploadClient from "@/components/apps/UploadClient";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export const dynamic = "force-dynamic";

export default async function NewVersionPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  return (
    <DashboardShell title="Publish Update">
      <Link href={`/apps/${id}`} className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowLeft className="mr-1 h-4 w-4" /> Back to app
      </Link>
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Upload &amp; Sign</CardTitle>
          <CardDescription>Upload a new IPA version. This triggers the signing engine automatically.</CardDescription>
        </CardHeader>
        <CardContent>
          <UploadClient appId={id} />
        </CardContent>
      </Card>
    </DashboardShell>
  );
}
