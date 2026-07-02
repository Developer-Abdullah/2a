import Foundation

class AppRepository {
    func fetchApps(cursor: String?) async throws -> CursorPage<AppItem> {
        let endpoint = cursor != nil ? "/v1/apps?cursor=\(cursor!)" : "/v1/apps"
        return try await APIService.shared.request(endpoint: endpoint)
    }
}