import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import SendNotificationForm from "@/components/notifications/SendNotificationForm";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export const dynamic = "force-dynamic";

export default function NewNotificationPage() {
  return (
    <DashboardShell title="Broadcast Notification">
      <Link href="/notifications" className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowLeft className="mr-1 h-4 w-4" /> Back to notifications
      </Link>
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Send a broadcast</CardTitle>
          <CardDescription>Delivered to all activated devices in the store.</CardDescription>
        </CardHeader>
        <CardContent>
          <SendNotificationForm />
        </CardContent>
      </Card>
    </DashboardShell>
  );
}
