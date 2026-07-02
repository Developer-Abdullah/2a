// Manual payment details shown to the customer on a pending order. Values come from env so they can
// be set without a code change; sensible Arabic placeholders are used when unset.
export interface PaymentInfo {
  instructions: string;
  whatsapp: string; // digits only, for wa.me links; empty hides the WhatsApp button
}

export function paymentInfo(currency: string): PaymentInfo {
  const whatsapp = (process.env.NEXT_PUBLIC_WHATSAPP || "").replace(/[^0-9]/g, "");
  const egp = process.env.NEXT_PUBLIC_PAY_EGP || "فودافون كاش / إنستاباي: (أضف رقم الدفع من إعدادات المتجر)";
  const kwd = process.env.NEXT_PUBLIC_PAY_KWD || "تحويل بنكي / K-NET: (أضف بيانات الدفع من إعدادات المتجر)";
  return { instructions: currency === "KWD" ? kwd : egp, whatsapp };
}
