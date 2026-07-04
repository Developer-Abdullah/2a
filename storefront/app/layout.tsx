import type { Metadata } from "next";
import "./globals.css";
import { StoreProvider } from "@/components/store-provider";
import { Header } from "@/components/header";
import { Footer } from "@/components/footer";
import { getCurrency } from "@/lib/api";
import { getLocale } from "@/lib/locale-server";
import { dir } from "@/lib/i18n";

export const metadata: Metadata = {
  title: "Double A — أكواد اشتراك تطبيقات بلس",
  description: "متجر Double A: أكواد تفعيل فورية لتطبيقات بلس للآيفون والآيباد. تفعيل سريع، ضمان، ودعم على مدار الساعة.",
  icons: { icon: "/logo.jpg" },
};

// Applies the stored theme before paint to avoid a flash of the wrong theme.
const noFlashTheme = `(function(){try{var t=localStorage.getItem('store_theme');if(t==='dark'||(!t&&window.matchMedia('(prefers-color-scheme: dark)').matches)){document.documentElement.classList.add('dark');}}catch(e){}})();`;

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const [currency, locale] = await Promise.all([getCurrency(), getLocale()]);
  return (
    <html lang={locale} dir={dir(locale)} suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: noFlashTheme }} />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700;800&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="min-h-screen font-sans">
        <StoreProvider initialCurrency={currency} initialLocale={locale}>
          <Header />
          <main>{children}</main>
          <Footer />
        </StoreProvider>
      </body>
    </html>
  );
}
