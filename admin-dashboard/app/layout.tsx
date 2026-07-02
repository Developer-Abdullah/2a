import "./globals.css";
import type { Metadata } from "next";
import AuthProvider from "@/components/providers/AuthProvider";

export const metadata: Metadata = {
  title: "Platform Admin",
  description: "Admin dashboard for the app distribution platform",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-slate-50 text-slate-900 antialiased">
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
