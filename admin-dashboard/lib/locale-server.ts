import { cookies } from "next/headers";
import { DEFAULT_LOCALE, LOCALE_COOKIE, Locale, translate } from "@/lib/i18n";

// Reads the active locale from the cookie for server components. Next 16's cookies() is async.
export async function getLocale(): Promise<Locale> {
  const store = await cookies();
  const v = store.get(LOCALE_COOKIE)?.value;
  return v === "en" || v === "ar" ? v : DEFAULT_LOCALE;
}

// Server-side translator bound to the request's locale: `const t = await getT(); t("key")`.
export async function getT(): Promise<(key: string) => string> {
  const locale = await getLocale();
  return (key: string) => translate(locale, key);
}
