"use client";

import { signOut } from "next-auth/react";
import { LogOut, Moon, Sun, Languages } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useTheme } from "@/components/providers/ThemeProvider";
import { useI18n } from "@/components/providers/LocaleProvider";

// Header shows the page title plus the language + theme toggles and sign out. `title` may be an i18n
// key (translated) or a plain string (rendered as-is via the dictionary fallback).
export default function Header({ title, email }: { title: string; email?: string | null }) {
  const { theme, toggle: toggleTheme } = useTheme();
  const { t, locale, toggle: toggleLocale } = useI18n();

  return (
    <header className="flex h-16 items-center justify-between border-b border-border bg-card px-6 md:px-8">
      <h1 className="text-xl font-semibold text-foreground">{t(title)}</h1>
      <div className="flex items-center gap-2">
        {email ? <span className="hidden text-sm text-muted-foreground sm:inline">{email}</span> : null}

        <Button
          variant="outline"
          size="sm"
          onClick={toggleLocale}
          title={t("header.language")}
          aria-label={t("header.language")}
        >
          <Languages className="h-4 w-4" />
          <span className="ms-1.5 font-semibold">{locale === "ar" ? "EN" : "ع"}</span>
        </Button>

        <Button
          variant="outline"
          size="icon"
          onClick={toggleTheme}
          title={theme === "dark" ? t("header.theme.toLight") : t("header.theme.toDark")}
          aria-label={theme === "dark" ? t("header.theme.toLight") : t("header.theme.toDark")}
        >
          {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
        </Button>

        <Button variant="outline" size="sm" onClick={() => signOut({ callbackUrl: "/login" })}>
          <LogOut className="h-4 w-4" />
          <span className="ms-1.5 hidden sm:inline">{t("header.signout")}</span>
        </Button>
      </div>
    </header>
  );
}
