import Foundation
import Security

class KeychainService {
    static let shared = KeychainService()
    private let tokenKey = "com.platform.accessToken"
    private let deviceKey = "com.platform.deviceId"

    func saveAccessToken(_ token: String) { save(token, account: tokenKey) }
    func getAccessToken() -> String? { read(account: tokenKey) }

    /// The device fingerprint chosen at activation. Must be sent as X-Device-ID on every
    /// authenticated request so it matches the device_id baked into the access token.
    func saveDeviceId(_ id: String) { save(id, account: deviceKey) }
    func getDeviceId() -> String? { read(account: deviceKey) }

    private func save(_ value: String, account: String) {
        let data = value.data(using: .utf8)!
        let query: [String: Any] = [kSecClass as String: kSecClassGenericPassword, kSecAttrAccount as String: account, kSecValueData as String: data, kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlock]
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }

    private func read(account: String) -> String? {
        let query: [String: Any] = [kSecClass as String: kSecClassGenericPassword, kSecAttrAccount as String: account, kSecReturnData as String: kCFBooleanTrue!, kSecMatchLimit as String: kSecMatchLimitOne]
        var dataTypeRef: AnyObject?
        if SecItemCopyMatching(query as CFDictionary, &dataTypeRef) == errSecSuccess, let data = dataTypeRef as? Data {
            return String(data: data, encoding: .utf8)
        }
        return nil
    }
}
