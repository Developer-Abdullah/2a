import SwiftUI

struct UpdateBannerView: View {
    let update: Update
    var body: some View {
        VStack {
            Text("Update Required").font(.headline).foregroundColor(.white)
            Text("A mandatory update is available.").font(.subheadline).foregroundColor(.white.opacity(0.8))
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color.red)
    }
}