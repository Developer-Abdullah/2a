import Foundation

@MainActor
class ActivationViewModel: ObservableObject {
    @Published var isActivating = false
    @Published var error: String?
    
    func activate(code: String) async {
        isActivating = true
        // Call ActivationRepository
        isActivating = false
    }
}