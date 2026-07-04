import Foundation

// Decodable models matching the core API's JSON (snake_case is converted by the decoder).

struct Price: Decodable, Hashable {
    let currency: String
    let amount: Double
    let compareAt: Double?
}

struct Product: Decodable, Identifiable, Hashable {
    let id: String
    let slug: String
    let name: String
    let subtitle: String?
    let description: String?
    let deviceType: String
    let features: [String]
    let imageUrl: String?
    let purchaseCount: Int
    let ratingAvg: Double
    let ratingCount: Int
    let prices: [Price]

    func price(_ currency: String) -> Price? {
        prices.first { $0.currency == currency } ?? prices.first
    }
}

struct CodeStatus: Decodable {
    let valid: Bool
    let deviceType: String
    let maxDevices: Int
    let currentDeviceCount: Int
    let isRevoked: Bool
    let expired: Bool
}

struct InstallableApp: Decodable {
    let ready: Bool
    let name: String?
    let version: String?
    let manifestUrl: String?
}
