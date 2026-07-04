import SwiftUI

@main
struct DoubleAApp: App {
    @StateObject private var state = AppState()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(state)
                .environment(\.theme, state.theme)
                .environment(\.layoutDirection, state.layoutDirection)
                .preferredColorScheme(state.dark ? .dark : .light)
                .tint(state.theme.primary)
        }
    }
}

struct RootView: View {
    @EnvironmentObject var state: AppState
    @Environment(\.theme) var theme

    var body: some View {
        TabView {
            HomeView()
                .tabItem { Label(state.t("tab.home"), systemImage: "square.grid.2x2.fill") }
            ActivateView()
                .tabItem { Label(state.t("tab.activate"), systemImage: "key.fill") }
            SettingsView()
                .tabItem { Label(state.t("tab.settings"), systemImage: "gearshape.fill") }
        }
        .background(theme.background.ignoresSafeArea())
    }
}
