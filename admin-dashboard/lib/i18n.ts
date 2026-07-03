// Lightweight bilingual (Arabic / English) dictionary for the admin dashboard. `t(locale, key)`
// returns the translation, falling back to the key itself so untranslated strings (e.g. legacy page
// titles passed as literals) still render.

export type Locale = "ar" | "en";
export const LOCALES: Locale[] = ["ar", "en"];
export const DEFAULT_LOCALE: Locale = "ar";
export const LOCALE_COOKIE = "da_locale";

type Dict = Record<string, string>;

const ar: Dict = {
  "brand.subtitle": "لوحة إدارة المتجر",
  "header.signout": "تسجيل الخروج",
  "header.theme.toDark": "الوضع الليلي",
  "header.theme.toLight": "الوضع النهاري",
  "header.language": "English",

  "nav.dashboard": "لوحة التحكم",
  "nav.products": "المنتجات",
  "nav.orders": "الطلبات",
  "nav.apps": "التطبيقات",
  "nav.codes": "الأكواد",
  "nav.ratings": "التقييمات",
  "nav.notifications": "الإشعارات",
  "nav.customers": "العملاء",
  "nav.devices": "الأجهزة",
  "nav.analytics": "الإحصائيات",
  "nav.enrollment": "التسجيل",
  "nav.settings": "الإعدادات",

  "dash.sales": "المبيعات",
  "dash.store": "المتجر",
  "dash.revenue_egp": "إيرادات (جنيه)",
  "dash.revenue_kwd": "إيرادات (دينار)",
  "dash.orders_today": "طلبات اليوم",
  "dash.orders_month": "طلبات هذا الشهر",
  "dash.pending": "بانتظار المراجعة",
  "dash.fulfilled": "طلبات مكتملة",
  "dash.codes_issued": "أكواد صادرة",
  "dash.best_seller": "الأكثر مبيعًا",
  "dash.products": "المنتجات/التطبيقات",
  "dash.customers": "المستخدمون",
  "dash.devices": "الأجهزة النشطة",
  "dash.avg_rating": "متوسط التقييم",
  "dash.load_error": "تعذّر تحميل الإحصائيات",

  "products.title": "المنتجات",
  "products.new": "منتج جديد",
  "products.count": "منتج",
  "products.col.name": "الاسم",
  "products.col.id": "المُعرّف",
  "products.col.prices": "الأسعار",
  "products.col.sales": "المبيعات",
  "products.col.status": "الحالة",
  "products.published": "منشور",
  "products.draft": "مسودة",

  "orders.title": "الطلبات",
  "orders.count": "طلب",
  "orders.col.email": "البريد",
  "orders.col.total": "الإجمالي",
  "orders.col.codes": "الأكواد",
  "orders.col.status": "الحالة",
  "orders.status.pending": "بانتظار الدفع",
  "orders.status.paid": "مدفوع",
  "orders.status.failed": "ملغى",
  "orders.status.fulfilled": "مكتمل",

  "common.back": "رجوع",
  "common.edit": "تعديل",
  "common.loading": "جارٍ التحميل…",
};

const en: Dict = {
  "brand.subtitle": "Store admin",
  "header.signout": "Sign out",
  "header.theme.toDark": "Dark mode",
  "header.theme.toLight": "Light mode",
  "header.language": "عربي",

  "nav.dashboard": "Dashboard",
  "nav.products": "Products",
  "nav.orders": "Orders",
  "nav.apps": "Apps",
  "nav.codes": "Codes",
  "nav.ratings": "Ratings",
  "nav.notifications": "Notifications",
  "nav.customers": "Customers",
  "nav.devices": "Devices",
  "nav.analytics": "Analytics",
  "nav.enrollment": "Enrollment",
  "nav.settings": "Settings",

  "dash.sales": "Sales",
  "dash.store": "Store",
  "dash.revenue_egp": "Revenue (EGP)",
  "dash.revenue_kwd": "Revenue (KWD)",
  "dash.orders_today": "Orders today",
  "dash.orders_month": "Orders this month",
  "dash.pending": "Pending review",
  "dash.fulfilled": "Fulfilled orders",
  "dash.codes_issued": "Codes issued",
  "dash.best_seller": "Best seller",
  "dash.products": "Products / apps",
  "dash.customers": "Customers",
  "dash.devices": "Active devices",
  "dash.avg_rating": "Average rating",
  "dash.load_error": "Failed to load stats",

  "products.title": "Products",
  "products.new": "New product",
  "products.count": "products",
  "products.col.name": "Name",
  "products.col.id": "Slug",
  "products.col.prices": "Prices",
  "products.col.sales": "Sales",
  "products.col.status": "Status",
  "products.published": "Published",
  "products.draft": "Draft",

  "orders.title": "Orders",
  "orders.count": "orders",
  "orders.col.email": "Email",
  "orders.col.total": "Total",
  "orders.col.codes": "Codes",
  "orders.col.status": "Status",
  "orders.status.pending": "Awaiting payment",
  "orders.status.paid": "Paid",
  "orders.status.failed": "Cancelled",
  "orders.status.fulfilled": "Fulfilled",

  "common.back": "Back",
  "common.edit": "Edit",
  "common.loading": "Loading…",
};

const dicts: Record<Locale, Dict> = { ar, en };

export function translate(locale: Locale, key: string): string {
  return dicts[locale]?.[key] ?? dicts[DEFAULT_LOCALE][key] ?? key;
}

export function dir(locale: Locale): "rtl" | "ltr" {
  return locale === "ar" ? "rtl" : "ltr";
}
