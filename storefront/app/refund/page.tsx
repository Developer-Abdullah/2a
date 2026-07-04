import { PolicyPage, PolicySection } from "@/components/policy";
import { getLocale, getT } from "@/lib/locale-server";

export const metadata = { title: "Refund — Double A" };

const SECTIONS: Record<"ar" | "en", PolicySection[]> = {
  ar: [
    { heading: "طبيعة المنتج الرقمي", body: [
      "نظرًا لأن المنتجات عبارة عن أكواد رقمية تُسلَّم فورًا، فإن الكود غير قابل للاسترجاع بعد إصداره أو تفعيله.",
    ] },
    { heading: "الحالات المؤهلة للاسترجاع", body: [
      "إذا لم يُصدَر الكود بعد (الطلب ما زال قيد المراجعة ولم يُسلَّم أي كود)، يمكن إلغاء الطلب واسترجاع المبلغ.",
      "إذا كان الكود غير صالح أو لا يعمل لسبب من طرفنا، نستبدله بكود جديد أو نعيد المبلغ.",
    ] },
    { heading: "الضمان", body: [
      "يشمل الضمان استبدال الكود في حال وجود مشكلة تقنية في التفعيل خلال فترة الضمان الموضحة في صفحة المنتج.",
    ] },
    { heading: "كيفية طلب الاسترجاع", body: [
      "تواصل معنا عبر قنوات التواصل مع ذكر رقم الطلب، وسنراجع حالتك ونرد عليك في أقرب وقت.",
    ] },
  ],
  en: [
    { heading: "Digital product nature", body: [
      "Since the products are digital codes delivered instantly, a code is non-refundable once it has been issued or activated.",
    ] },
    { heading: "Refund-eligible cases", body: [
      "If the code hasn't been issued yet (the order is still under review and no code was delivered), the order can be cancelled and refunded.",
      "If the code is invalid or doesn't work due to a fault on our side, we replace it with a new code or refund you.",
    ] },
    { heading: "Warranty", body: [
      "The warranty covers replacing the code if there's a technical activation problem within the warranty period shown on the product page.",
    ] },
    { heading: "How to request a refund", body: [
      "Contact us through our channels with your order number, and we'll review your case and reply as soon as possible.",
    ] },
  ],
};

export default async function RefundPage() {
  const [locale, t] = await Promise.all([getLocale(), getT()]);
  return <PolicyPage title={t("footer.refund")} updated="2026-07-04" sections={SECTIONS[locale]} />;
}
