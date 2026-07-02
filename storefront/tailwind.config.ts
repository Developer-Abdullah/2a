import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // Double A brand: deep navy primary (logo background top).
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
        // Double A accent: the crimson red from the logo's lower gradient.
        accent: {
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
        ink: "#101827",
      },
      fontFamily: {
        sans: ["var(--font-cairo)", "system-ui", "sans-serif"],
      },
      boxShadow: {
        card: "0 10px 30px -12px rgba(19, 28, 51, 0.22)",
      },
    },
  },
  plugins: [],
};
export default config;
