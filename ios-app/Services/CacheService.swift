import Foundation

class CacheService {
    static let shared = CacheService()
    private let fileManager = FileManager.default
    
    func save<T: Codable>(_ object: T, key: String) {
        // Disk caching implementation
    }
    
    func load<T: Codable>(key: String) -> T? {
        return nil
    }
}