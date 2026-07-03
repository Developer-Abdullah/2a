"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";

type Theme = "light" | "dark";
const ThemeContext = createContext<{ theme: Theme; toggle: () => void; setTheme: (t: Theme) => void }>({
  theme: "light",
  toggle: () => {},
  setTheme: () => {},
});

export const THEME_KEY = "da_theme";

// Applies/removes the `dark` class on <html> and persists the choice. An inline script in the root
// layout applies the stored theme before paint to avoid a flash.
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>("light");

  useEffect(() => {
    const stored = (typeof window !== "undefined" && (localStorage.getItem(THEME_KEY) as Theme)) || null;
    const initial: Theme = stored || (document.documentElement.classList.contains("dark") ? "dark" : "light");
    setThemeState(initial);
  }, []);

  const apply = useCallback((t: Theme) => {
    setThemeState(t);
    if (typeof window !== "undefined") {
      localStorage.setItem(THEME_KEY, t);
      document.documentElement.classList.toggle("dark", t === "dark");
    }
  }, []);

  const toggle = useCallback(() => apply(theme === "dark" ? "light" : "dark"), [theme, apply]);

  return <ThemeContext.Provider value={{ theme, toggle, setTheme: apply }}>{children}</ThemeContext.Provider>;
}

export const useTheme = () => useContext(ThemeContext);
