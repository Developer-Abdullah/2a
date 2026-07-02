import Foundation

struct User: Codable, Identifiable {
    let id: String
    let activationCodeId: String
    let displayName: String
    let createdAt: Date
    let lastSeenAt: Date
}