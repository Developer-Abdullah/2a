import Foundation

class NotificationRepository {
    func fetchHistory() async throws -> CursorPage<NotificationItem> {
        return try await APIService.shared.request(endpoint: "/v1/notifications")
    }
}