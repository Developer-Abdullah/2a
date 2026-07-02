import Foundation

struct AppItem: Codable, Identifiable {
    let id: String
    let name: String
    let iconUrl: String
    let bundleIdentifier: String
    let minimumOsVersion: String
    let category: String
    let description: String
    let features: [String]
    let screenshotUrls: [String]
    let ipaManifestUrl: String
    let uploadedAt: Date
    let isNew: Bool
    let latestVersion: AppVersion
}