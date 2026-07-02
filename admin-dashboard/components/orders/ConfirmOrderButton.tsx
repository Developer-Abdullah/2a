"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { CheckCircle2 } from "lucide-react";
import { confirmOrderAction } from "@/lib/actions";
import { Button } from "@/components/ui/button";

// ConfirmOrderButton confirms an offline (manual) payment: it mints the codes and refreshes so the
// admin sees them immediately.
export default function ConfirmOrderButton({ id }: { id: string }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function confirm() {
    setPending(true);
    setError("");
    const res = await confirmOrderAction(id);
    setPending(false);
    if (res.ok) {
      router.refresh();
    } else {
      setError(res.error || "تعذّر تأكيد الطلب");
    }
  }

  return (
    <div className="space-y-2">
      <Button onClick={confirm} disabled={pending} className="w-full">
        <CheckCircle2 className="ml-1 h-4 w-4" />
        {pending ? "جارٍ التأكيد…" : "تأكيد الدفع وإصدار الأكواد"}
      </Button>
      {error ? <p className="text-xs text-red-600">{error}</p> : null}
    </div>
  );
}
