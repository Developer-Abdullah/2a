import SwiftUI

// Global app configuration + user preferences (language, theme). Preferences persist in UserDefaults.
// Set `apiBaseURL` to your API origin (e.g. https://api.2a-plus.com) — no trailing /v1.
final class AppState: ObservableObject {
    // MARK: Config
    let apiBaseURL = "https://api.2a-plus.com"
    let storeSlug = "store"
    let whatsappNumber = "201557583049"   // digits only

    // MARK: Preferences
    @Published var lang: Lang {
        didSet { UserDefaults.standard.set(lang.rawValue, forKey: "lang") }
    }
    @Published var dark: Bool {
        didSet { UserDefaults.standard.set(dark, forKey: "dark") }
    }

    var isRTL: Bool { lang == .ar }
    var layoutDirection: LayoutDirection { isRTL ? .rightToLeft : .leftToRight }
    var theme: Theme { Theme(dark: dark) }

    init() {
        let saved = UserDefaults.standard.string(forKey: "lang")
        self.lang = Lang(rawValue: saved ?? "ar") ?? .ar
        if UserDefaults.standard.object(forKey: "dark") != nil {
            self.dark = UserDefaults.standard.bool(forKey: "dark")
        } else {
            self.dark = false
        }
    }

    func t(_ key: String) -> String { L(key, lang) }

    // A stable per-install device id (so re-activations from this device don't burn extra slots).
    var deviceId: String {
        if let id = UserDefaults.standard.string(forKey: "device_id") { return id }
        let id = UUID().uuidString
        UserDefaults.standard.set(id, forKey: "device_id")
        return id
    }

    func deviceLabel(_ type: String) -> String {
        switch type { case "iphone": return t("device.iphone"); case "ipad": return t("device.ipad"); default: return t("device.both") }
    }

    var whatsappURL: URL? { URL(string: "https://wa.me/\(whatsappNumber)") }
}
