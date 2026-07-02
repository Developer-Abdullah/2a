import Foundation

@MainActor
class RatingsViewModel: ObservableObject {
    @Published var summary: AppRatingSummary?
    
    func fetchSummary() async {
        // Call RatingRepository
    }
    
    func submitRating(_ rating: Int, comment: String) async {
        // POST to RatingRepository
    }
}