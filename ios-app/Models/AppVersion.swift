import Foundation

struct AppVersion: Codable, Identifiable {
    let id: String
    let version: String
    let buildNumber: String
    let releaseNotes: String?
    let releasedAt: Date
    let sizeMb: Double
    let downloadCount: Int
    let signingStatus: SigningStatus
    let updateType: UpdateType?
}