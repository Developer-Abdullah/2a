import Foundation

enum SigningStatus: String, Codable {
    case pending
    case signing
    case signed
    case failed
}