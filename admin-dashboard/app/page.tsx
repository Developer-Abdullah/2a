import { getServerSession } from "next-auth";
import { redirect } from "next/navigation";
import { authOptions } from "@/lib/auth";

export const dynamic = "force-dynamic";

// Single-store mode: every authenticated admin lands on the store dashboard. The old owner/merchant
// split is retired — the platform runs one storefront now.
export default async function RootPage() {
  const session = await getServerSession(authOptions);
  if (!session) redirect("/login");
  redirect("/dashboard");
}
