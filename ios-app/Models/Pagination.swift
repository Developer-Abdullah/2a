import Foundation

struct PaginationMeta: Codable {
    let nextCursor: String?
    let hasMore: Bool
    let total: Int?
}

struct CursorPage<T: Codable>: Codable {
    let items: [T]
    let pagination: PaginationMeta
}