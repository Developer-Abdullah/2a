import { redirect } from "next/navigation";

// Retired: merchant onboarding no longer applies in single-store mode.
export default function NewMerchantPage() {
  redirect("/dashboard");
}
