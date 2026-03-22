import SwiftUI

enum Tab: String, CaseIterable, Identifiable {
    case chat
    case decisions
    case search
    case settings

    var id: String { rawValue }

    var label: String {
        switch self {
        case .chat: return "Chat"
        case .decisions: return "Decisions"
        case .search: return "Search"
        case .settings: return "Settings"
        }
    }

    var icon: String {
        switch self {
        case .chat: return "bubble.left.and.bubble.right"
        case .decisions: return "doc.text"
        case .search: return "magnifyingglass"
        case .settings: return "gearshape"
        }
    }
}

@Observable
final class AppState {
    var decisions: [Decision] = []
    var selectedTab: Tab = .chat
    var searchQuery: String = ""
    var isLoading: Bool = false
    var repoPath: String = ""
    var apiKey: String = ""
    var chatMessages: [ChatMessage] = []
    var isTyping: Bool = false
    var lastSynced: Date? = nil
    var selectedDecision: Decision? = nil
    var decisionFilter: DecisionFilter = .all
    var showAddDecision: Bool = false

    enum DecisionFilter: String, CaseIterable {
        case all = "All"
        case product = "Product"
        case technical = "Technical"
        case stale = "Stale"
    }

    var filteredDecisions: [Decision] {
        switch decisionFilter {
        case .all:
            return decisions
        case .product:
            return decisions.filter { $0.type.lowercased() == "product" }
        case .technical:
            return decisions.filter { $0.type.lowercased() == "technical" }
        case .stale:
            return decisions.filter { $0.stale }
        }
    }

    var searchResults: [Decision] {
        guard !searchQuery.isEmpty else { return [] }
        let query = searchQuery.lowercased()
        return decisions.filter { decision in
            decision.title.lowercased().contains(query) ||
            decision.body.lowercased().contains(query) ||
            decision.author.lowercased().contains(query) ||
            decision.anchor.lowercased().contains(query)
        }
    }

    func loadSettings() {
        if let saved = UserDefaults.standard.string(forKey: "repoPath") {
            repoPath = saved
        }
        if let saved = UserDefaults.standard.string(forKey: "apiKey") {
            apiKey = saved
        }
    }

    func saveSettings() {
        UserDefaults.standard.set(repoPath, forKey: "repoPath")
        UserDefaults.standard.set(apiKey, forKey: "apiKey")
    }
}
