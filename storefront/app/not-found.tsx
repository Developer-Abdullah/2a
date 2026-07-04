import Link from "next/link";
import { getLocale, getT } from "@/lib/locale-server";

export default async function NotFound() {
  const [locale, t] = await Promise.all([getLocale(), getT()]);
  return (
    <div className="container-page py-24 text-center">
      <h1 className="text-5xl font-extrabold text-brand-600 dark:text-brand-300">404</h1>
      <p className="mt-4 text-lg text-muted-foreground">{locale === "ar" ? "الصفحة غير موجودة." : "Page not found."}</p>
      <Link href="/" className="btn-primary mt-8">{t("order.back_home")}</Link>
    </div>
  );
}
