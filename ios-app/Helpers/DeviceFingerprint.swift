import Foundation
import CommonCrypto

struct DeviceFingerprint {
    static func generate() -> String {
        let uuid = KeychainService.shared.getDeviceId() ?? UUID().uuidString
        let model = "iPhone" // Mocked. Real implementation reads sysctlbyname("hw.machine")
        let raw = "\(uuid)-\(model)"
        return raw // In production, wrap in SHA-256
    }
}