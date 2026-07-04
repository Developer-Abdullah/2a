import type { Locale } from "./i18n";

// Manual payment methods shown to the customer on a pending order, per currency. Each method has a
// copyable transfer number. WhatsApp (for sending payment proof) comes from env.
export interface PaymentMethod {
  name: string;
  number: string;
}

export interface PaymentInfo {
  methods: PaymentMethod[];
  whatsapp: string; // digits only, for wa.me links; empty hides the WhatsApp button
}

const EGP_METHODS = (locale: Locale): PaymentMethod[] => [
  { name: locale === "ar" ? "تحويل إنستاباي" : "InstaPay transfer", number: "01040450058" },
  { name: locale === "ar" ? "تحويل فودافون كاش" : "Vodafone Cash transfer", number: "01040821796" },
];

const KWD_METHODS = (locale: Locale): PaymentMethod[] => [
  { name: locale === "ar" ? "تحويل عن طريق ومض" : "Wamd transfer", number: "55245607" },
];

export function paymentInfo(currency: string, locale: Locale = "ar"): PaymentInfo {
  const whatsapp = (process.env.NEXT_PUBLIC_WHATSAPP || "").replace(/[^0-9]/g, "");
  return {
    methods: currency === "KWD" ? KWD_METHODS(locale) : EGP_METHODS(locale),
    whatsapp,
  };
}
