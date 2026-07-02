import Foundation

enum APIError: Error, LocalizedError {
    case invalidURL
    case server(code: String, message: String)
    case http(status: Int)
    case decoding(Error)

    var errorDescription: String? {
        switch self {
        case .invalidURL: return "Invalid URL"
        case .server(_, let message): return message
        case .http(let status): return "Request failed (\(status))"
        case .decoding(let error): return "Failed to decode response: \(error.localizedDescription)"
        }
    }
}

/// The backend wraps every response in { "success": Bool, "data": T?, "error": {code, message}? }.
private struct Envelope<T: Decodable>: Decodable {
    let success: Bool
    let data: T?
    let error: APIErrorModel?
}

/// Generic networking client for the Core API. Endpoints already include their version prefix
/// (e.g. "/v1/apps"), so `baseURL` is the bare origin with NO trailing "/v1".
final class APIService {
    static let shared = APIService()

    /// Override per build/tenant, e.g. "https://store.platform.com".
    var baseURL = "https://platform.com"
    /// Sent as X-Tenant-Slug so the API can resolve the tenant without subdomain routing.
    var tenantSlug = "store"

    private let decoder: JSONDecoder = {
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        decoder.dateDecodingStrategy = .custom { d in
            let container = try d.singleValueContainer()
            let raw = try container.decode(String.self)
            if let date = APIService.parseDate(raw) { return date }
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Unrecognized date format: \(raw)")
        }
        return decoder
    }()

    /// Performs a request and returns the decoded `data` payload, unwrapping the success envelope.
    func request<T: Decodable>(endpoint: String, method: String = "GET", body: Data? = nil) async throws -> T {
        guard let url = URL(string: baseURL + endpoint) else { throw APIError.invalidURL }

        var req = URLRequest(url: url)
        req.httpMethod = method
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.setValue(tenantSlug, forHTTPHeaderField: "X-Tenant-Slug")
        if let token = KeychainService.shared.getAccessToken() {
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        if let deviceId = KeychainService.shared.getDeviceId() {
            req.setValue(deviceId, forHTTPHeaderField: "X-Device-ID")
        }
        req.httpBody = body

        let (data, response) = try await URLSession.shared.data(for: req)
        let status = (response as? HTTPURLResponse)?.statusCode ?? 0

        let envelope: Envelope<T>
        do {
            envelope = try decoder.decode(Envelope<T>.self, from: data)
        } catch {
            if !(200...299).contains(status) { throw APIError.http(status: status) }
            throw APIError.decoding(error)
        }

        if !envelope.success || !(200...299).contains(status) {
            if let err = envelope.error { throw APIError.server(code: err.code, message: err.message) }
            throw APIError.http(status: status)
        }
        guard let payload = envelope.data else {
            throw APIError.decoding(DecodingError.valueNotFound(T.self,
                .init(codingPath: [], debugDescription: "Response had no data")))
        }
        return payload
    }

    /// Accepts both ISO8601 and the PostgreSQL timestamp text the API currently emits.
    static func parseDate(_ raw: String) -> Date? {
        let iso = ISO8601DateFormatter()
        iso.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = iso.date(from: raw) { return date }
        iso.formatOptions = [.withInternetDateTime]
        if let date = iso.date(from: raw) { return date }

        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = TimeZone(identifier: "UTC")
        for format in ["yyyy-MM-dd HH:mm:ss.SSSSSSxxx", "yyyy-MM-dd HH:mm:ssxxx", "yyyy-MM-dd HH:mm:ss", "yyyy-MM-dd"] {
            formatter.dateFormat = format
            if let date = formatter.date(from: raw) { return date }
        }
        return nil
    }
}
