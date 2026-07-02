import Foundation

class ActivationRepository {
    func validate(code: String, deviceId: String) async throws -> ActivationResponse {
        let body = try JSONEncoder().encode(["code": code, "device_id": deviceId, "device_type": DeviceTypeDetector.getDeviceType()])
        return try await APIService.shared.request(endpoint: "/v1/activation/validate", method: "POST", body: body)
    }
}