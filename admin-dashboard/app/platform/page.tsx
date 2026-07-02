import { redirect } from "next/navigation";

// Retired: the cross-tenant platform view no longer applies in single-store mode.
export default function PlatformPage() {
  redirect("/dashboard");
}
