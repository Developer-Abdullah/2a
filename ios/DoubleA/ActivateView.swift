import SwiftUI

struct ActivateView: View {
    @EnvironmentObject var state: AppState
    @Environment(\.theme) var theme

    enum Step { case code, activate, done }

    @State private var code = ""
    @State private var step: Step = .code
    @State private var status: CodeStatus?
    @State private var checking = false
    @State private var activating = false
    @State private var message: String?     // inline error/status text
    @State private var app: InstallableApp?
    @State private var loadingApp = false

    private var api: API { API(baseURL: state.apiBaseURL, slug: state.storeSlug) }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    icon
                    Text(state.t("activate.title")).font(.title2.bold()).foregroundStyle(theme.text)
                    Text(state.t("activate.sub")).font(.subheadline).foregroundStyle(theme.muted)
                        .multilineTextAlignment(.center)

                    stepBar

                    codeField
                    if let message { Text(message).font(.subheadline).foregroundStyle(.red).multilineTextAlignment(.center) }
                    if step == .activate, let s = status, s.valid { activateCard(s) }
                    if step == .done { installCard }
                }
                .padding()
            }
            .background(theme.background.ignoresSafeArea())
            .navigationBarTitleDisplayMode(.inline)
        }
    }

    private var icon: some View {
        Image(systemName: "key.fill").font(.title).foregroundStyle(.white)
            .frame(width: 56, height: 56).background(Brand.gradient).clipShape(RoundedRectangle(cornerRadius: 16))
    }

    private var stepBar: some View {
        HStack(spacing: 8) {
            dot(1, step == .code, step != .code)
            line
            dot(2, step == .activate, step == .done)
            line
            dot(3, step == .done, false)
        }
    }
    private var line: some View { Rectangle().fill(theme.border).frame(width: 24, height: 1) }
    private func dot(_ n: Int, _ active: Bool, _ done: Bool) -> some View {
        Text("\(n)").font(.caption2.bold()).foregroundStyle(.white)
            .frame(width: 24, height: 24)
            .background(done ? Color.green : (active ? theme.primary : theme.muted)).clipShape(Circle())
    }

    private var codeField: some View {
        HStack {
            TextField(state.t("activate.placeholder"), text: $code)
                .multilineTextAlignment(.center).font(.system(.title3, design: .monospaced))
                .autocorrectionDisabled().textInputAutocapitalization(.characters)
                .padding().background(theme.card).clipShape(RoundedRectangle(cornerRadius: 12))
                .overlay(RoundedRectangle(cornerRadius: 12).stroke(theme.border, lineWidth: 1))
                .foregroundStyle(theme.text)
            Button(action: { Task { await check() } }) {
                if checking { ProgressView().tint(.white) } else { Text(state.t("activate.check")).bold() }
            }
            .padding(.horizontal, 18).frame(height: 52)
            .background(theme.primary).foregroundStyle(.white).clipShape(RoundedRectangle(cornerRadius: 12))
            .disabled(checking || code.isEmpty)
        }
    }

    private func activateCard(_ s: CodeStatus) -> some View {
        VStack(spacing: 14) {
            HStack {
                Image(systemName: "checkmark.circle.fill").foregroundStyle(.green)
                VStack(alignment: .leading) {
                    Text(state.t("activate.valid")).bold().foregroundStyle(theme.text)
                    Text("\(state.t("activate.devices")): \(s.currentDeviceCount)/\(s.maxDevices)")
                        .font(.caption).foregroundStyle(theme.muted)
                }
                Spacer()
            }
            Button(action: { Task { await activate() } }) {
                HStack { if activating { ProgressView().tint(.white) } ; Text(activating ? state.t("activate.activating") : state.t("activate.activate_btn")).bold() }
                    .frame(maxWidth: .infinity).padding().background(theme.primary).foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }.disabled(activating)
        }
        .padding().background(theme.card).clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(RoundedRectangle(cornerRadius: 16).stroke(theme.border, lineWidth: 1))
    }

    private var installCard: some View {
        VStack(spacing: 14) {
            HStack { Image(systemName: "checkmark.circle.fill").foregroundStyle(.green); Text(state.t("activate.done")).bold().foregroundStyle(theme.text); Spacer() }
            Divider()
            Label(state.t("activate.install_title"), systemImage: "iphone.and.arrow.forward").font(.headline).foregroundStyle(theme.text)
            if loadingApp {
                ProgressView()
            } else if let app, app.ready, let manifest = app.manifestUrl,
                      let url = URL(string: "itms-services://?action=download-manifest&url=\(manifest.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? manifest)") {
                Text(state.t("activate.install_body")).font(.subheadline).foregroundStyle(theme.muted)
                Link(destination: url) {
                    Label("\(state.t("activate.install_btn")) \(app.name ?? "")", systemImage: "square.and.arrow.down")
                        .frame(maxWidth: .infinity).padding().background(theme.primary).foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            } else {
                Text(state.t("activate.not_ready")).font(.subheadline).foregroundStyle(theme.muted)
                if let wa = state.whatsappURL {
                    Link(state.t("settings.whatsapp"), destination: wa).buttonStyle(.borderedProminent)
                }
            }
        }
        .padding().background(theme.card).clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(RoundedRectangle(cornerRadius: 16).stroke(theme.border, lineWidth: 1))
    }

    // MARK: Actions
    private func check() async {
        checking = true; message = nil; status = nil; step = .code
        do {
            let s = try await api.codeStatus(code.trimmingCharacters(in: .whitespaces))
            status = s
            if s.valid { step = .activate }
            else { message = s.isRevoked ? state.t("activate.revoked") : s.expired ? state.t("activate.expired") : state.t("activate.used") }
        } catch APIError.notFound {
            message = state.t("activate.notfound")
        } catch {
            message = error.localizedDescription
        }
        checking = false
    }

    private func activate() async {
        activating = true; message = nil
        do {
            let type = status?.deviceType == "ipad" ? "ipad" : "iphone"
            try await api.activate(code: code.trimmingCharacters(in: .whitespaces), deviceId: state.deviceId, deviceType: type)
            step = .done
            await loadApp()
        } catch {
            message = state.t("activate.err")
        }
        activating = false
    }

    private func loadApp() async {
        loadingApp = true
        app = try? await api.installableApp()
        loadingApp = false
    }
}
