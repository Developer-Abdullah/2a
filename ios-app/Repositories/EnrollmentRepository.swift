import Foundation

class EnrollmentRepository {
    func pollStatus(token: String) async throws -> EnrollmentSessionResponse {
        return try await APIService.shared.request(endpoint: "/v1/enroll/status?token=\(token)")
    }
}