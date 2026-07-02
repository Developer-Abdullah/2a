import { getServerSession } from "next-auth";
import { redirect } from "next/navigation";
import { authOptions } from "@/lib/auth";
import Sidebar from "./Sidebar";
import Header from "./Header";

// DashboardShell wraps every authenticated page: it enforces the session server-side (redirecting
// to /login when absent) and renders the persistent sidebar + header chrome.
export default async function DashboardShell({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  const session = await getServerSession(authOptions);
  if (!session) redirect("/login");

  return (
    <div className="flex min-h-screen bg-slate-50">
      <Sidebar isOwner={!!session.user?.isOwner} />
      <div className="flex min-w-0 flex-1 flex-col">
        <Header title={title} email={session.user?.email} />
        <main className="flex-1 p-8">{children}</main>
      </div>
    </div>
  );
}
