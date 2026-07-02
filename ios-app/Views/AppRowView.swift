import SwiftUI

struct AppRowView: View {
    let app: AppItem
    
    var body: some View {
        HStack {
            AsyncImage(url: URL(string: app.iconUrl)) { image in
                image.resizable().scaledToFill()
            } placeholder: {
                Color.gray.opacity(0.3)
            }
            .frame(width: 60, height: 60)
            .cornerRadius(12)
            
            VStack(alignment: .leading) {
                Text(app.name).font(.headline)
                Text("Version \(app.latestVersion.version)").font(.caption).foregroundColor(.secondary)
            }
            Spacer()
            Button("Get") {
                InstallationManager.installApp(manifestURL: app.ipaManifestUrl)
            }
            .padding(.horizontal, 16).padding(.vertical, 6)
            .background(Color.blue.opacity(0.1)).foregroundColor(.blue).clipShape(Capsule())
        }
        .padding(.horizontal)
    }
}