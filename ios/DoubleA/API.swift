import Foundation

// Self-contained networking client for the Double A core API. Unwraps the standard
// { success, data, error } envelope and adds the store tenant header.
enum APIError: LocalizedError {
    case badURL, notFound, server(String), http(Int), decode
    var errorDescription: String? {
        switch self {
        case .badURL: return "Bad URL"
        case .notFound: return "Not found"
        case .server(let m): return m
        case .http(let s): return "HTTP \(s)"
        case .decode: return "Decode error"
        }
    }
}

private struct Envelope<T: Decodable>: Decodable {
    let success: Bool
    let data: T?
    struct Err: Decodable { let code: String; let message: String }
    let error: Err?
}

struct API {
    let baseURL: String
    let slug: String

    private var decoder: JSONDecoder {
        let d = JSONDecoder()
        d.keyDecodingStrategy = .convertFromSnakeCase
        return d
    }

    private func get<T: Decodable>(_ path: String) async throws -> T {
        guard let url = URL(string: baseURL + path) else { throw APIError.badURL }
        var req = URLRequest(url: url)
        req.setValue(slug, forHTTPHeaderField: "X-Tenant-Slug")
        let (data, resp) = try await URLSession.shared.data(for: req)
        let status = (resp as? HTTPURLResponse)?.statusCode ?? 0
        if status == 404 { throw APIError.notFound }
        guard let env = try? decoder.decode(Envelope<T>.self, from: data) else { throw APIError.decode }
        if let payload = env.data, env.success { return payload }
        if let e = env.error { throw APIError.server(e.message) }
        throw APIError.http(status)
    }

    private func post<T: Decodable>(_ path: String, _ body: [String: Any]) async throws -> T {
        guard let url = URL(string: baseURL + path) else { throw APIError.badURL }
        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        req.setValue(slug, forHTTPHeaderField: "X-Tenant-Slug")
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try JSONSerialization.data(withJSONObject: body)
        let (data, resp) = try await URLSession.shared.data(for: req)
        let status = (resp as? HTTPURLResponse)?.statusCode ?? 0
        guard let env = try? decoder.decode(Envelope<T>.self, from: data) else { throw APIError.decode }
        if let payload = env.data, env.success { return payload }
        if let e = env.error { throw APIError.server(e.message) }
        throw APIError.http(status)
    }

    // MARK: Endpoints
    struct ProductsResponse: Decodable { let items: [Product] }
    func products(currency: String) async throws -> [Product] {
        let r: ProductsResponse = try await get("/shop/products?currency=\(currency)")
        return r.items
    }

    struct StatusResponse: Decodable { let status: CodeStatus }
    func codeStatus(_ code: String) async throws -> CodeStatus {
        let r: StatusResponse = try await get("/shop/activation?code=\(code.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? code)")
        return r.status
    }

    struct ActivateResponse: Decodable { let accessToken: String? }
    func activate(code: String, deviceId: String, deviceType: String) async throws {
        let _: ActivateResponse = try await post("/v1/activation/validate",
            ["code": code, "device_id": deviceId, "device_type": deviceType])
    }

    func installableApp() async throws -> InstallableApp {
        try await get("/shop/app")
    }
}
