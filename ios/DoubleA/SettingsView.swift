import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var state: AppState
    @Environment(\.theme) var theme

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 16) {
                    brandHeader

                    // Language
                    card {
                        row(icon: "globe", title: state.t("settings.language"))
                        Picker("", selection: Binding(get: { state.lang }, set: { state.lang = $0 })) {
                            Text("العربية").tag(Lang.ar)
                            Text("English").tag(Lang.en)
                        }.pickerStyle(.segmented)
                    }

                    // Theme
                    card {
                        row(icon: "circle.lefthalf.filled", title: state.t("settings.theme"))
                        Picker("", selection: Binding(get: { state.dark }, set: { state.dark = $0 })) {
                            Text(state.t("settings.theme.light")).tag(false)
                            Text(state.t("settings.theme.dark")).tag(true)
                        }.pickerStyle(.segmented)
                    }

                    // Contact
                    card {
                        row(icon: "bubble.left.and.bubble.right.fill", title: state.t("settings.contact"))
                        if let wa = state.whatsappURL {
                            Link(destination: wa) {
                                Label(state.t("settings.whatsapp"), systemImage: "phone.fill")
                                    .frame(maxWidth: .infinity).padding(.vertical, 10)
                                    .background(theme.primary).foregroundStyle(.white).clipShape(RoundedRectangle(cornerRadius: 10))
                            }
                        }
                    }

                    // About
                    card {
                        row(icon: "info.circle.fill", title: state.t("settings.about"))
                        Text(state.t("settings.about_body")).font(.subheadline).foregroundStyle(theme.muted)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }
                }
                .padding()
            }
            .background(theme.background.ignoresSafeArea())
            .navigationTitle(state.t("settings.title"))
            .navigationBarTitleDisplayMode(.inline)
        }
    }

    private var brandHeader: some View {
        VStack(spacing: 8) {
            Image(systemName: "a.circle.fill").font(.system(size: 44)).foregroundStyle(.white)
            Text("Double A").font(.title2.bold()).foregroundStyle(.white)
        }
        .frame(maxWidth: .infinity).padding(.vertical, 24)
        .background(Brand.gradient).clipShape(RoundedRectangle(cornerRadius: 20))
    }

    private func card<Content: View>(@ViewBuilder _ content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 12) { content() }
            .padding().frame(maxWidth: .infinity, alignment: .leading)
            .background(theme.card).clipShape(RoundedRectangle(cornerRadius: 16))
            .overlay(RoundedRectangle(cornerRadius: 16).stroke(theme.border, lineWidth: 1))
    }

    private func row(icon: String, title: String) -> some View {
        HStack(spacing: 10) {
            Image(systemName: icon).foregroundStyle(theme.primary)
            Text(title).font(.headline).foregroundStyle(theme.text)
            Spacer()
        }
    }
}
