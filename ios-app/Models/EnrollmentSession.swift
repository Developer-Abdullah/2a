import Foundation

struct EnrollmentSessionResponse: Codable {
    let completed: Bool
    let deviceType: String?
}