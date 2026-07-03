import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: { DEFAULT: "hsl(var(--primary))", foreground: "hsl(var(--primary-foreground))" },
        secondary: { DEFAULT: "hsl(var(--secondary))", foreground: "hsl(var(--secondary-foreground))" },
        destructive: { DEFAULT: "hsl(var(--destructive))", foreground: "hsl(var(--destructive-foreground))" },
        muted: { DEFAULT: "hsl(var(--muted))", foreground: "hsl(var(--muted-foreground))" },
        popover: { DEFAULT: "hsl(var(--popover))", foreground: "hsl(var(--popover-foreground))" },
        card: { DEFAULT: "hsl(var(--card))", foreground: "hsl(var(--card-foreground))" },
        // Double A brand: deep navy primary + crimson accent (mirrors the storefront).
        brand: {
          50: "#f1f4fa",
          100: "#e0e7f4",
          200: "#c4d1e8",
          300: "#9db2d6",
          400: "#6f8cc0",
          500: "#4d6dab",
          600: "#3b5695",
          700: "#31467a",
          800: "#2b3b64",
          900: "#1e2a4a",
          950: "#131c33",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
          50: "#fdf3f4",
          100: "#fbe5e7",
          200: "#f6cdd1",
          300: "#eea6ad",
          400: "#e37581",
          500: "#d34a5b",
          600: "#b93547",
          700: "#a02f3f",
          800: "#822a38",
          900: "#6f2833",
          950: "#3d1119",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
    },
  },
  plugins: [],
};
export default config;
