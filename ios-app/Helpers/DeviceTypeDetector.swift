import UIKit

struct DeviceTypeDetector {
    static func getDeviceType() -> String {
        return UIDevice.current.userInterfaceIdiom == .pad ? "ipad" : "iphone"
    }
}