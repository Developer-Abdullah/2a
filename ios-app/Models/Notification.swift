import Foundation

struct NotificationItem: Codable, Identifiable {
    let id: String
    let title: String
    let body: String
    let targetAppId: String?
    let sentAt: Date
    let readAt: Date?
}