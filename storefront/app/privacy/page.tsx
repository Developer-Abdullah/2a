import { PolicyPage, PolicySection } from "@/components/policy";
import { getLocale, getT } from "@/lib/locale-server";

export const metadata = { title: "Privacy — Double A" };

const SECTIONS: Record<"ar" | "en", PolicySection[]> = {
  ar: [
    { heading: "البيانات التي نجمعها", body: [
      "نجمع الحد الأدنى من البيانات اللازمة لإتمام طلبك: البريد الإلكتروني ورقم الجوال (اختياري) وتفاصيل الطلب.",
      "لا نقوم بتخزين بيانات بطاقات الدفع؛ الدفع يتم يدويًا عبر الوسائل الموضحة.",
    ] },
    { heading: "كيفية استخدام البيانات", body: [
      "نستخدم بريدك لإرسال كود التفعيل وتحديثات حالة الطلب، ورقم جوالك للتواصل بخصوص الدفع عند الحاجة.",
      "لا نبيع بياناتك أو نشاركها مع أطراف ثالثة لأغراض تسويقية.",
    ] },
    { heading: "التقييمات", body: [
      "عند إرسال تقييم، نخزّن التقييم والتعليق بشكل مجهول، مع بصمة مشفّرة لعنوان الإنترنت لمنع التكرار فقط.",
    ] },
    { heading: "التواصل", body: [
      "لأي استفسار حول بياناتك أو لطلب حذفها، تواصل معنا عبر القنوات الموجودة في تذييل الصفحة.",
    ] },
  ],
  en: [
    { heading: "Data we collect", body: [
      "We collect the minimum data needed to complete your order: email, phone (optional), and order details.",
      "We do not store payment card data; payment is made manually via the shown methods.",
    ] },
    { heading: "How we use data", body: [
      "We use your email to send the activation code and order status updates, and your phone to contact you about payment when needed.",
      "We do not sell your data or share it with third parties for marketing.",
    ] },
    { heading: "Reviews", body: [
      "When you submit a review, we store the rating and comment anonymously, with a hashed IP fingerprint used only to prevent duplicates.",
    ] },
    { heading: "Contact", body: [
      "For any question about your data or to request its deletion, contact us through the channels in the page footer.",
    ] },
  ],
};

export default async function PrivacyPage() {
  const [locale, t] = await Promise.all([getLocale(), getT()]);
  return <PolicyPage title={t("footer.privacy")} updated="2026-07-04" sections={SECTIONS[locale]} />;
}
