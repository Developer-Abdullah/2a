import type { Metadata } from "next";
import "./globals.css";
import { StoreProvider } from "@/components/store-provider";
import { Header } from "@/components/header";
import { Footer } from "@/components/footer";
import { getCurrency } from "@/lib/api";

export const metadata: Metadata = {
  title: "Double A — أكواد اشتراك تطبيقات بلس",
  description: "متجر Double A: أكواد تفعيل فورية لتطبيقات بلس للآيفون والآيباد. تفعيل سريع، ضمان، ودعم على مدار الساعة.",
  icons: { icon: "/logo.jpg" },
};

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const currency = await getCurrency();
  return (
    <html lang="ar" dir="rtl">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700;800&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="min-h-screen font-sans">
        <StoreProvider initialCurrency={currency}>
          <Header />
          <main>{children}</main>
          <Footer />
        </StoreProvider>
      </body>
    </html>
  );
}
