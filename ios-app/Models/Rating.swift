import Foundation

struct AppRating: Codable {
    let rating: Int
    let comment: String?
    let createdAt: Date
}

struct AppRatingSummary: Codable {
    let average: Double
    let count: Int
    let breakdown: [String: Int]
}