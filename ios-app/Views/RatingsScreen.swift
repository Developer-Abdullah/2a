import SwiftUI

struct RatingsScreen: View {
    @State private var rating = 5
    @State private var comment = ""
    var body: some View {
        Form {
            Section(header: Text("Rate the Store")) {
                Stepper("Stars: \(rating)", value: $rating, in: 1...5)
                TextField("Add a comment...", text: $comment)
            }
            Button("Submit Rating") { /* Call ViewModel */ }
        }
        .navigationTitle("Feedback")
    }
}