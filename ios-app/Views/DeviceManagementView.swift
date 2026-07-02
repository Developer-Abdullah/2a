import SwiftUI

struct DeviceManagementView: View {
    var body: some View {
        List {
            Text("iPhone 15 Pro").badge("Current")
            Text("iPad Pro").swipeActions {
                Button("Release", role: .destructive) { /* Call ViewModel */ }
            }
        }
        .navigationTitle("My Devices")
    }
}