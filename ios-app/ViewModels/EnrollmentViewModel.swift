import Foundation
import UIKit
import Combine

// Represents the precise state machine of the OTA Enrollment flow
enum EnrollmentState: Equatable {
    case idle
    case loadingToken
    case downloading(token: String)
    case polling(token: String)
    case success(deviceId: String)
    case error(String)
}

// Mock definition for the API Response
struct EnrollmentStatusResponse: Codable {
    let completed: Bool
    let deviceId: String?
    let deviceType: String?
}

@MainActor
final class EnrollmentViewModel: ObservableObject {
    @Published var state: EnrollmentState = .idle
    
    private var pollTask: Task<Void, Never>?
    
    // In a real app, this would be injected. We use a mock API wrapper for demonstration.
    private let tenantSlug = "my-store"
    
    /// Starts the flow by acquiring a session token, then handing off to Safari
    func downloadProfile() async {
        state = .loadingToken
        
        do {
            // Step 1: Ask backend to generate an enrollment session and token
            // Note: As per architecture, we assume the API provides the token before Safari redirect.
            let token = try await fetchEnrollmentSessionToken()
            
            // Transition state before opening Safari so `scenePhase` logic knows we have a token
            state = .downloading(token: token)
            
            // Step 2: Open Safari to trigger the mobileconfig download
            // Safari is REQUIRED; in-app WKWebView blocks configuration profile installations.
            let urlString = "https://\(tenantSlug).platform.com/v1/enroll?token=\(token)"
            guard let url = URL(string: urlString) else {
                state = .error("Invalid enrollment URL")
                return
            }
            
            UIApplication.shared.open(url, options: [:], completionHandler: nil)
            
        } catch {
            state = .error(error.localizedDescription)
        }
    }
    
    /// Recursively polls the backend every 3 seconds to check if Apple's callback has fired
    func startPolling(token: String) {
        // Cancel any existing task to prevent duplicate network spam
        pollTask?.cancel()
        state = .polling(token: token)
        
        pollTask = Task {
            while !Task.isCancelled {
                do {
                    let status = try await checkEnrollmentStatus(token: token)
                    
                    if status.completed, let deviceId = status.deviceId {
                        self.state = .success(deviceId: deviceId)
                        break // Exit the polling loop
                    }
                } catch {
                    // We catch and ignore transient network errors (e.g., poor signal)
                    // to keep the polling loop alive until timeout or success.
                    print("Polling transient error: \(error.localizedDescription)")
                }
                
                // Sleep for exactly 3 seconds before the next poll
                try? await Task.sleep(nanoseconds: 3_000_000_000)
            }
        }
    }
    
    /// Called by the View when the App returns to the foreground (.active scene phase)
    func resumePollingIfNeeded() {
        switch state {
        case .downloading(let token):
            // User returned from Safari/Settings. Begin polling.
            startPolling(token: token)
        case .polling(let token):
            // Ensure the polling task wasn't suspended/killed by iOS jetsam
            startPolling(token: token)
        default:
            break
        }
    }
    
    func cancelPolling() {
        pollTask?.cancel()
        pollTask = nil
        if case .polling = state {
            state = .idle
        }
    }
    
    // MARK: - API Mock Implementations
    
    private func fetchEnrollmentSessionToken() async throws -> String {
        // Simulate network latency for token generation
        try await Task.sleep(nanoseconds: 1_000_000_000)
        return UUID().uuidString.replacingOccurrences(of: "-", with: "")
    }
    
    private func checkEnrollmentStatus(token: String) async throws -> EnrollmentStatusResponse {
        guard let url = URL(string: "https://\(tenantSlug).platform.com/v1/enroll/status?token=\(token)") else {
            throw URLError(.badURL)
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpRes = response as? HTTPURLResponse, httpRes.statusCode == 200 else {
            throw URLError(.badServerResponse)
        }
        
        // Ensure snake_case mapping is used per backend JSON tags
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        return try decoder.decode(EnrollmentStatusResponse.self, from: data)
    }
}