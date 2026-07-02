import Foundation

enum UpdateType: String, Codable {
    case optional
    case forced
    case silent
}

struct Update: Codable, Identifiable {
    let id: String
    let appId: String
    let currentVersion: String
    let latestVersion: String
    let updateType: UpdateType
}