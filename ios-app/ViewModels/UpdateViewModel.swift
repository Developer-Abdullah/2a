import Foundation

@MainActor
class UpdateViewModel: ObservableObject {
    @Published var forcedUpdateApp: AppItem?
    
    func checkForUpdates() async {
        // Evaluate VersionComparator and UpdateService
    }
}