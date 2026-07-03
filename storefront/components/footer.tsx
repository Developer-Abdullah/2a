import Link from "next/link";
import { MessageCircle, Send, Instagram } from "lucide-react";

// Social/contact links. Instagram is fixed; WhatsApp and Telegram come from env so they can be set
// without a code change. A link is only rendered when its target is configured.
const INSTAGRAM = "https://www.instagram.com/2a_mtjr";

function whatsappLink(): string {
  const n = (process.env.NEXT_PUBLIC_WHATSAPP || "").replace(/[^0-9]/g, "");
  return n ? `https://wa.me/${n}` : "";
}

function telegramLink(): string {
  const t = (process.env.NEXT_PUBLIC_TELEGRAM || "").replace(/^@/, "");
  return t ? `https://t.me/${t}` : "";
}

export function Footer() {
  const wa = whatsappLink();
  const tg = telegramLink();
  return (
    <footer className="mt-20 border-t border-slate-200 bg-white">
      <div className="container-page grid gap-8 py-12 md:grid-cols-3">
        <div>
          <div className="flex items-center gap-2">
            <img src="/logo.jpg" alt="Double A" className="h-9 w-9 rounded-full" />
            <h3 className="text-lg font-extrabold text-brand-800">Double A</h3>
          </div>
          <p className="mt-3 text-sm leading-6 text-slate-500">
            أكواد اشتراك فورية لتطبيقات بلس. تفعيل سريع، ضمان، ودعم على مدار الساعة.
          </p>
        </div>
        <div>
          <h4 className="font-bold text-ink">روابط</h4>
          <ul className="mt-3 space-y-2 text-sm text-slate-500">
            <li><Link href="/" className="hover:text-brand-700">الرئيسية</Link></li>
            <li><Link href="/#products" className="hover:text-brand-700">الباقات</Link></li>
            <li><Link href="/orders" className="hover:text-brand-700">تتبّع طلباتي</Link></li>
          </ul>
        </div>
        <div>
          <h4 className="font-bold text-ink">المتجر</h4>
          <ul className="mt-3 space-y-2 text-sm text-slate-500">
            <li><Link href="/terms" className="hover:text-brand-700">الشروط والأحكام</Link></li>
            <li><Link href="/refund" className="hover:text-brand-700">سياسة الاسترجاع</Link></li>
            <li><Link href="/privacy" className="hover:text-brand-700">سياسة الخصوصية</Link></li>
          </ul>
          <div className="mt-4 flex items-center gap-3">
            <a href={INSTAGRAM} target="_blank" rel="noopener noreferrer" aria-label="إنستجرام" className="rounded-xl bg-brand-50 p-2 text-brand-700 hover:bg-brand-100"><Instagram className="h-5 w-5" /></a>
            {wa && <a href={wa} target="_blank" rel="noopener noreferrer" aria-label="واتساب" className="rounded-xl bg-brand-50 p-2 text-brand-700 hover:bg-brand-100"><MessageCircle className="h-5 w-5" /></a>}
            {tg && <a href={tg} target="_blank" rel="noopener noreferrer" aria-label="تيليجرام" className="rounded-xl bg-brand-50 p-2 text-brand-700 hover:bg-brand-100"><Send className="h-5 w-5" /></a>}
          </div>
        </div>
      </div>
      <div className="border-t border-slate-100 py-4 text-center text-xs text-slate-400">
        © {new Date().getFullYear()} Double A. جميع الحقوق محفوظة.
      </div>
    </footer>
  );
}
