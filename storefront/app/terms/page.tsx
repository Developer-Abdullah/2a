import { PolicyPage, PolicySection } from "@/components/policy";
import { getLocale, getT } from "@/lib/locale-server";

export const metadata = { title: "Terms — Double A" };

const SECTIONS: Record<"ar" | "en", PolicySection[]> = {
  ar: [
    { heading: "طبيعة المنتجات", body: [
      "متجر Double A يبيع أكواد اشتراك رقمية للتطبيقات. المنتج عبارة عن كود تفعيل يُسلَّم إلكترونيًا بعد تأكيد الدفع.",
      "الكود صالح لعملية تفعيل واحدة ولمدة الاشتراك الموضحة في صفحة المنتج.",
    ] },
    { heading: "الدفع والتسليم", body: [
      "يتم الدفع يدويًا عبر الوسائل المتاحة في صفحة الطلب. بعد استلام المبلغ والتحقق منه، يُصدر الكود ويظهر في صفحة الطلب ويُرسل إلى بريدك.",
      "يُرجى الاحتفاظ برقم الطلب كمرجع عند التواصل بخصوص الدفع.",
    ] },
    { heading: "مسؤولية المستخدم", body: [
      "أنت مسؤول عن إدخال بريد إلكتروني صحيح لاستلام الكود، وعن الحفاظ على سرية الكود بعد استلامه.",
      "لا يجوز إعادة بيع الأكواد أو مشاركتها بطريقة تخالف شروط الاستخدام.",
    ] },
    { heading: "التعديلات", body: [
      "قد نقوم بتحديث هذه الشروط من وقت لآخر. استمرارك في استخدام المتجر يعني موافقتك على النسخة المحدثة.",
    ] },
  ],
  en: [
    { heading: "Nature of the products", body: [
      "Double A sells digital app subscription codes. The product is an activation code delivered electronically after payment is confirmed.",
      "A code is valid for a single activation and for the subscription period shown on the product page.",
    ] },
    { heading: "Payment & delivery", body: [
      "Payment is made manually via the methods shown on the order page. Once the amount is received and verified, the code is issued, shown on the order page, and emailed to you.",
      "Please keep your order number as a reference when contacting us about payment.",
    ] },
    { heading: "User responsibility", body: [
      "You are responsible for entering a correct email to receive the code, and for keeping the code confidential after receiving it.",
      "Reselling or sharing codes in a way that violates the terms of use is not allowed.",
    ] },
    { heading: "Changes", body: [
      "We may update these terms from time to time. Your continued use of the store means you accept the updated version.",
    ] },
  ],
};

export default async function TermsPage() {
  const [locale, t] = await Promise.all([getLocale(), getT()]);
  return <PolicyPage title={t("footer.terms")} updated="2026-07-04" sections={SECTIONS[locale]} />;
}
