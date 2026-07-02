import Link from "next/link";
import { MessageCircle, Send, Instagram } from "lucide-react";

export function Footer() {
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
            <li><Link href="/#reviews" className="hover:text-brand-700">آراء العملاء</Link></li>
          </ul>
        </div>
        <div>
          <h4 className="font-bold text-ink">تواصل معنا</h4>
          <div className="mt-3 flex items-center gap-3">
            <a href="#" aria-label="واتساب" className="rounded-xl bg-brand-50 p-2 text-brand-700 hover:bg-brand-100"><MessageCircle className="h-5 w-5" /></a>
            <a href="#" aria-label="تيليجرام" className="rounded-xl bg-brand-50 p-2 text-brand-700 hover:bg-brand-100"><Send className="h-5 w-5" /></a>
            <a href="#" aria-label="إنستجرام" className="rounded-xl bg-brand-50 p-2 text-brand-700 hover:bg-brand-100"><Instagram className="h-5 w-5" /></a>
          </div>
        </div>
      </div>
      <div className="border-t border-slate-100 py-4 text-center text-xs text-slate-400">
        © {new Date().getFullYear()} Double A. جميع الحقوق محفوظة.
      </div>
    </footer>
  );
}
