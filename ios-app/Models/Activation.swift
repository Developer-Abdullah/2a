import Foundation

struct ActivationResponse: Codable {
    let accessToken: String
    let refreshToken: String
    let userId: String
    let expiresIn: Int
}