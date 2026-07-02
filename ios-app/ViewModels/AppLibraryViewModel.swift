import Foundation

@MainActor
class AppLibraryViewModel: ObservableObject {
    @Published var apps: [AppItem] = []
    @Published var isLoading = false
    
    func fetchApps() async {
        isLoading = true
        // Integration with AppRepository goes here
        isLoading = false
    }
}