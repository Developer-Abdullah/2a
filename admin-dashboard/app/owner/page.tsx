import { redirect } from "next/navigation";

// Retired: the multi-merchant owner console no longer applies in single-store mode.
export default function OwnerPage() {
  redirect("/dashboard");
}
