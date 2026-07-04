import SwiftUI

// Double A brand colors + light/dark theme tokens. Mirrors the web storefront (navy + crimson).
extension Color {
    init(hex: String) {
        let s = hex.trimmingCharacters(in: CharacterSet(charactersIn: "#"))
        var v: UInt64 = 0
        Scanner(string: s).scanHexInt64(&v)
        let r, g, b: Double
        if s.count == 6 {
            r = Double((v >> 16) & 0xff) / 255
            g = Double((v >> 8) & 0xff) / 255
            b = Double(v & 0xff) / 255
        } else { r = 0; g = 0; b = 0 }
        self.init(.sRGB, red: r, green: g, blue: b, opacity: 1)
    }
}

enum Brand {
    static let navy = Color(hex: "#31467a")     // brand-700
    static let navyDeep = Color(hex: "#131c33")  // brand-950
    static let navyLight = Color(hex: "#4d6dab")  // brand-500
    static let crimson = Color(hex: "#b93547")    // accent-600
    static let amber = Color(hex: "#f5b301")

    static let gradient = LinearGradient(
        colors: [Color(hex: "#131c33"), Color(hex: "#2b3b64"), Color(hex: "#5d2c48"), Color(hex: "#a02f3f")],
        startPoint: .topLeading, endPoint: .bottomTrailing
    )
}

// Theme-aware surfaces so light/dark both look right.
struct Theme {
    let dark: Bool
    var background: Color { dark ? Color(hex: "#0b1220") : Color(hex: "#f8fafc") }
    var card: Color { dark ? Color(hex: "#131c33") : .white }
    var text: Color { dark ? Color(hex: "#e8edf6") : Color(hex: "#101827") }
    var muted: Color { dark ? Color(hex: "#93a4c3") : Color(hex: "#64748b") }
    var border: Color { dark ? Color(hex: "#2b3b64") : Color(hex: "#e2e8f0") }
    var primary: Color { dark ? Brand.navyLight : Brand.navy }
}

private struct ThemeKey: EnvironmentKey {
    static let defaultValue = Theme(dark: false)
}
extension EnvironmentValues {
    var theme: Theme {
        get { self[ThemeKey.self] }
        set { self[ThemeKey.self] = newValue }
    }
}
