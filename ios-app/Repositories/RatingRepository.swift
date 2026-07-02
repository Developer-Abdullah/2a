import Foundation

class RatingRepository {
    func fetchSummary() async throws -> AppRatingSummary {
        return try await APIService.shared.request(endpoint: "/v1/ratings/summary")
    }
}