import SwiftUI

@main
struct BackstoryApp: App {
    @State private var appState = AppState()

    var body: some Scene {
        WindowGroup {
            ContentView(state: appState)
                .frame(minWidth: 900, minHeight: 600)
                .preferredColorScheme(.dark)
                .background(Theme.background)
        }
        .windowStyle(.hiddenTitleBar)
        .defaultSize(width: 1100, height: 750)
    }
}
