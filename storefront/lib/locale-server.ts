import { cookies } from "next/headers";
import { DEFAULT_LOCALE, LOCALE_COOKIE, Locale, translate } from "./i18n";

// Reads the active locale from the cookie for server components (Next 16 cookies() is async).
export async function getLocale(): Promise<Locale> {
  const v = (await cookies()).get(LOCALE_COOKIE)?.value;
  return v === "en" || v === "ar" ? v : DEFAULT_LOCALE;
}

// Server-side translator bound to the request locale: `const t = await getT(); t("key")`.
export async function getT(): Promise<(key: string) => string> {
  const locale = await getLocale();
  return (key: string) => translate(locale, key);
}
