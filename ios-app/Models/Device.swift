import Foundation

struct Device: Codable, Identifiable {
    let id: String
    let deviceType: String
    let modelString: String
    let isRevoked: Bool
    let lastValidatedAt: Date
}