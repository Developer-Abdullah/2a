import SwiftUI

struct ErrorView: View {
    let message: String
    let retryAction: () -> Void
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle.fill").font(.largeTitle).foregroundColor(.red)
            Text(message).multilineTextAlignment(.center).foregroundColor(.secondary)
            Button("Try Again", action: retryAction).buttonStyle(.borderedProminent)
        }
        .padding()
    }
}