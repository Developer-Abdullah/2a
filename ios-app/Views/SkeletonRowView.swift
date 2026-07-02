import SwiftUI

struct SkeletonRowView: View {
    var body: some View {
        HStack {
            RoundedRectangle(cornerRadius: 12).fill(Color.gray.opacity(0.2)).frame(width: 60, height: 60)
            VStack(alignment: .leading) {
                RoundedRectangle(cornerRadius: 4).fill(Color.gray.opacity(0.2)).frame(width: 120, height: 16)
                RoundedRectangle(cornerRadius: 4).fill(Color.gray.opacity(0.2)).frame(width: 80, height: 12)
            }
            Spacer()
            RoundedRectangle(cornerRadius: 16).fill(Color.gray.opacity(0.2)).frame(width: 60, height: 32)
        }
        .padding(.horizontal)
    }
}