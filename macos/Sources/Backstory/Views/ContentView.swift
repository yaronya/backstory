import SwiftUI

struct ContentView: View {
    @Bindable var state: AppState

    var body: some View {
        NavigationSplitView {
            SidebarView(state: state)
                .navigationSplitViewColumnWidth(min: 200, ideal: Theme.sidebarWidth, max: 280)
        } detail: {
            ZStack {
                Theme.background.ignoresSafeArea()

                Group {
                    switch state.selectedTab {
                    case .chat:
                        ChatView(state: state)
                    case .decisions:
                        DecisionListView(state: state)
                    case .search:
                        SearchView(state: state)
                    case .settings:
                        SettingsView(state: state)
                    }
                }
                .transition(.opacity)
            }
        }
        .background(Theme.background)
        .onAppear {
            state.loadSettings()
            loadDecisions()
        }
    }

    private func loadDecisions() {
        guard !state.repoPath.isEmpty else { return }
        Task {
            let service = BackstoryService(repoPath: state.repoPath)
            do {
                let loaded = try await service.loadAllDecisions()
                await MainActor.run {
                    withAnimation(Theme.springAnimation) {
                        state.decisions = loaded
                    }
                }
            } catch {
                // silently handle
            }
        }
    }
}
