import UIKit

class InstallationManager {
    static func installApp(manifestURL: String) {
        guard let encoded = manifestURL.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed),
              let url = URL(string: "itms-services://?action=download-manifest&url=\(encoded)") else { return }
        DispatchQueue.main.async { UIApplication.shared.open(url, options: [:]) }
    }
}
