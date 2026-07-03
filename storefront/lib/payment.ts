// Manual payment methods shown to the customer on a pending order, per currency. Each method has a
// copyable transfer number. WhatsApp (for sending payment proof) comes from env.
export interface PaymentMethod {
  name: string;
  number: string;
}

export interface PaymentInfo {
  methods: PaymentMethod[];
  note: string;
  whatsapp: string; // digits only, for wa.me links; empty hides the WhatsApp button
}

const EGP_METHODS: PaymentMethod[] = [
  { name: "تحويل إنستاباي", number: "01040450058" },
  { name: "تحويل فودافون كاش", number: "01040821796" },
];

const KWD_METHODS: PaymentMethod[] = [
  { name: "تحويل عن طريق ومض", number: "55245607" },
];

export function paymentInfo(currency: string): PaymentInfo {
  const whatsapp = (process.env.NEXT_PUBLIC_WHATSAPP || "").replace(/[^0-9]/g, "");
  const isKWD = currency === "KWD";
  return {
    methods: isKWD ? KWD_METHODS : EGP_METHODS,
    note: "حوّل المبلغ على أي رقم من الأرقام، ثم ارفع صورة إيصال التحويل بالأسفل أو أرسلها لنا على واتساب.",
    whatsapp,
  };
}
