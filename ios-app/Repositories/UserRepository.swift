import Foundation

class UserRepository {
    func fetchProfile() async throws -> User {
        return try await APIService.shared.request(endpoint: "/v1/profile")
    }
}