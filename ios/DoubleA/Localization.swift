import SwiftUI

enum Lang: String { case ar, en }

// Lightweight in-app localization so language switches instantly without an app restart. The active
// language lives in AppState; `L(_, lang)` looks the string up, falling back to the key.
private let strings: [Lang: [String: String]] = [
    .ar: [
        "app.name": "Double A",
        "tab.home": "الرئيسية",
        "tab.activate": "التفعيل",
        "tab.settings": "الإعدادات",

        "home.title": "الباقات",
        "home.subtitle": "أكواد اشتراك فورية لتطبيقات بلس",
        "home.loading": "جارٍ التحميل…",
        "home.empty": "لا توجد باقات حاليًا.",
        "home.buy_note": "اشترِ الكود من المتجر ثم فعّله هنا.",
        "home.instant": "تفعيل فوري",

        "activate.title": "تفعيل الاشتراك",
        "activate.sub": "أدخل كودك، فعّله على هذا الجهاز، ثم ثبّت التطبيق.",
        "activate.placeholder": "XXXX-XXXX-XXXX",
        "activate.check": "تحقّق",
        "activate.valid": "الكود صالح ✅",
        "activate.devices": "الأجهزة",
        "activate.notfound": "الكود غير موجود.",
        "activate.revoked": "هذا الكود موقوف.",
        "activate.expired": "انتهت صلاحية هذا الكود.",
        "activate.used": "تم استخدام هذا الكود بالكامل.",
        "activate.activate_btn": "فعّل على هذا الجهاز",
        "activate.activating": "جارٍ التفعيل…",
        "activate.err": "تعذّر تفعيل الكود. تأكد أنه صالح.",
        "activate.done": "تم التفعيل على جهازك 🎉",
        "activate.install_title": "ثبّت التطبيق",
        "activate.install_body": "اضغط للتثبيت ووافق على تثبيت التطبيق، وستجد بداخله مكتبة التطبيقات.",
        "activate.install_btn": "تثبيت",
        "activate.not_ready": "التطبيق قيد التجهيز حاليًا. تواصل معنا للحصول على رابط التثبيت.",

        "settings.title": "الإعدادات",
        "settings.language": "اللغة",
        "settings.theme": "المظهر",
        "settings.theme.light": "فاتح",
        "settings.theme.dark": "غامق",
        "settings.contact": "تواصل معنا",
        "settings.whatsapp": "واتساب",
        "settings.about": "عن المتجر",
        "settings.about_body": "Double A — أكواد اشتراك فورية لتطبيقات بلس. تفعيل سريع، ضمان، ودعم على مدار الساعة.",

        "common.retry": "إعادة المحاولة",
        "device.iphone": "آيفون", "device.ipad": "آيباد", "device.both": "آيفون / آيباد",
    ],
    .en: [
        "app.name": "Double A",
        "tab.home": "Home",
        "tab.activate": "Activate",
        "tab.settings": "Settings",

        "home.title": "Packages",
        "home.subtitle": "Instant subscription codes for Plus apps",
        "home.loading": "Loading…",
        "home.empty": "No packages right now.",
        "home.buy_note": "Buy the code from the store, then activate it here.",
        "home.instant": "Instant activation",

        "activate.title": "Activate subscription",
        "activate.sub": "Enter your code, activate it on this device, then install the app.",
        "activate.placeholder": "XXXX-XXXX-XXXX",
        "activate.check": "Check",
        "activate.valid": "Code is valid ✅",
        "activate.devices": "Devices",
        "activate.notfound": "Code not found.",
        "activate.revoked": "This code is revoked.",
        "activate.expired": "This code has expired.",
        "activate.used": "This code is fully used.",
        "activate.activate_btn": "Activate on this device",
        "activate.activating": "Activating…",
        "activate.err": "Couldn't activate the code. Make sure it's valid.",
        "activate.done": "Activated on your device 🎉",
        "activate.install_title": "Install the app",
        "activate.install_body": "Tap install and approve the installation — the app library is inside.",
        "activate.install_btn": "Install",
        "activate.not_ready": "The app is being prepared. Contact us for the install link.",

        "settings.title": "Settings",
        "settings.language": "Language",
        "settings.theme": "Appearance",
        "settings.theme.light": "Light",
        "settings.theme.dark": "Dark",
        "settings.contact": "Contact us",
        "settings.whatsapp": "WhatsApp",
        "settings.about": "About",
        "settings.about_body": "Double A — instant subscription codes for Plus apps. Fast activation, warranty, and 24/7 support.",

        "common.retry": "Retry",
        "device.iphone": "iPhone", "device.ipad": "iPad", "device.both": "iPhone / iPad",
    ],
]

func L(_ key: String, _ lang: Lang) -> String {
    strings[lang]?[key] ?? strings[.ar]?[key] ?? key
}
