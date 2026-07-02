"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

// While an order is pending, re-render the server component every few seconds so it flips to
// "fulfilled" once the payment webhook has been processed. Stops after a bounded number of tries so
// an abandoned checkout doesn't poll forever.
export function OrderPoller({ id }: { id: string }) {
  const router = useRouter();
  useEffect(() => {
    let tries = 0;
    const timer = setInterval(() => {
      tries += 1;
      if (tries > 40) {
        clearInterval(timer);
        return;
      }
      router.refresh();
    }, 4000);
    return () => clearInterval(timer);
  }, [id, router]);
  return null;
}
