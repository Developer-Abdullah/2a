import SwiftUI

struct HomeView: View {
    @EnvironmentObject var state: AppState
    @Environment(\.theme) var theme

    @State private var products: [Product] = []
    @State private var loading = true
    @State private var error: String?
    @State private var currency = "EGP"

    private var api: API { API(baseURL: state.apiBaseURL, slug: state.storeSlug) }

    var body: some View {
        NavigationStack {
            ScrollView {
                header
                if loading {
                    ProgressView(state.t("home.loading")).padding(40)
                } else if let error {
                    errorView(error)
                } else if products.isEmpty {
                    Text(state.t("home.empty")).foregroundStyle(theme.muted).padding(40)
                } else {
                    LazyVStack(spacing: 14) {
                        ForEach(products) { p in ProductRow(product: p, currency: currency) }
                    }
                    .padding(.horizontal)
                    .padding(.bottom, 24)
                }
            }
            .background(theme.background.ignoresSafeArea())
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .principal) {
                    Text(state.t("app.name")).font(.headline.bold()).foregroundStyle(theme.text)
                }
            }
        }
        .task { await load() }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(state.t("home.title")).font(.largeTitle.bold()).foregroundStyle(.white)
            Text(state.t("home.subtitle")).foregroundStyle(.white.opacity(0.9))
            Picker("", selection: $currency) {
                Text("EGP ج.م").tag("EGP"); Text("KWD د.ك").tag("KWD")
            }
            .pickerStyle(.segmented)
            .onChange(of: currency) { _ in Task { await load() } }
            .padding(.top, 6)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .padding(.bottom, 8)
        .background(Brand.gradient)
    }

    private func errorView(_ msg: String) -> some View {
        VStack(spacing: 12) {
            Text(msg).foregroundStyle(theme.muted).multilineTextAlignment(.center)
            Button(state.t("common.retry")) { Task { await load() } }.buttonStyle(.borderedProminent)
        }.padding(40)
    }

    private func load() async {
        loading = true; error = nil
        do { products = try await api.products(currency: currency) }
        catch { self.error = error.localizedDescription }
        loading = false
    }
}

struct ProductRow: View {
    let product: Product
    let currency: String
    @Environment(\.theme) var theme
    @EnvironmentObject var state: AppState

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            ZStack(alignment: .topLeading) {
                Brand.gradient.frame(height: 96).clipShape(RoundedRectangle(cornerRadius: 14))
                HStack {
                    Label(state.t("home.instant"), systemImage: "bolt.fill")
                        .font(.caption.bold()).padding(.horizontal, 10).padding(.vertical, 5)
                        .background(Brand.amber).foregroundStyle(Brand.navyDeep).clipShape(Capsule())
                    Spacer()
                    Text(state.deviceLabel(product.deviceType))
                        .font(.caption.bold()).padding(.horizontal, 10).padding(.vertical, 5)
                        .background(.black.opacity(0.35)).foregroundStyle(.white).clipShape(Capsule())
                }.padding(10)
            }
            Text(product.name).font(.headline).foregroundStyle(theme.text)
            if let s = product.subtitle ?? product.description { Text(s).font(.subheadline).foregroundStyle(theme.muted).lineLimit(2) }
            HStack {
                Image(systemName: "star.fill").foregroundStyle(Brand.amber).font(.caption)
                Text(String(format: "%.2f", product.ratingAvg)).font(.caption.bold()).foregroundStyle(theme.text)
                Text("(\(product.ratingCount))").font(.caption).foregroundStyle(theme.muted)
                Spacer()
                if let price = product.price(currency) {
                    Text(money(price.amount, price.currency)).font(.title3.bold()).foregroundStyle(theme.primary)
                }
            }
        }
        .padding(14)
        .background(theme.card)
        .clipShape(RoundedRectangle(cornerRadius: 18))
        .overlay(RoundedRectangle(cornerRadius: 18).stroke(theme.border, lineWidth: 1))
    }

    private func money(_ v: Double, _ cur: String) -> String {
        let n = cur == "KWD" ? String(format: "%.3f", v) : String(format: "%.0f", v)
        return "\(n) \(cur)"
    }
}
