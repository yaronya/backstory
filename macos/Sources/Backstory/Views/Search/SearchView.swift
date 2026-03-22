import SwiftUI
import Combine

struct SearchView: View {
    @Bindable var state: AppState
    @FocusState private var isSearchFocused: Bool
    @State private var debounceTask: Task<Void, Never>? = nil

    var body: some View {
        VStack(spacing: 0) {
            header

            Divider().overlay(Theme.border)

            searchBar

            if state.searchQuery.isEmpty {
                emptyPrompt
            } else if state.searchResults.isEmpty {
                noResults
            } else {
                resultsList
            }
        }
        .background(Theme.background)
        .onAppear {
            isSearchFocused = true
        }
    }

    private var header: some View {
        HStack {
            Text("Search")
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)

            Spacer()

            if !state.searchQuery.isEmpty {
                Text("\(state.searchResults.count) results")
                    .font(.system(size: 12))
                    .foregroundStyle(Theme.textSecondary)
            }
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 14)
    }

    private var searchBar: some View {
        HStack(spacing: 10) {
            Image(systemName: "magnifyingglass")
                .font(.system(size: 14))
                .foregroundStyle(Theme.textSecondary)

            TextField("Search decisions...", text: $state.searchQuery)
                .textFieldStyle(.plain)
                .font(.system(size: 14))
                .foregroundStyle(Theme.textPrimary)
                .focused($isSearchFocused)

            if !state.searchQuery.isEmpty {
                Button {
                    withAnimation(Theme.springAnimation) {
                        state.searchQuery = ""
                    }
                } label: {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 14))
                        .foregroundStyle(Theme.textSecondary)
                }
                .buttonStyle(.plain)
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: Theme.cornerRadius)
                .fill(Theme.surfaceElevated)
        )
        .overlay(
            RoundedRectangle(cornerRadius: Theme.cornerRadius)
                .strokeBorder(isSearchFocused ? Theme.accent.opacity(0.5) : Theme.border, lineWidth: 1)
        )
        .padding(.horizontal, 20)
        .padding(.vertical, 12)
    }

    private var resultsList: some View {
        ScrollView {
            LazyVStack(spacing: 8) {
                ForEach(state.searchResults) { decision in
                    Button {
                        withAnimation(Theme.springAnimation) {
                            state.selectedDecision = decision
                        }
                    } label: {
                        DecisionRow(decision: decision)
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(20)
        }
        .sheet(item: $state.selectedDecision) { decision in
            DecisionDetailView(decision: decision)
        }
    }

    private var emptyPrompt: some View {
        VStack(spacing: 12) {
            Spacer()

            Image(systemName: "magnifyingglass")
                .font(.system(size: 40, weight: .thin))
                .foregroundStyle(Theme.textSecondary.opacity(0.4))

            Text("Search your team's decisions")
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            Text("Search by title, content, author, or anchor")
                .font(.system(size: 12))
                .foregroundStyle(Theme.textSecondary.opacity(0.7))

            Spacer()
        }
        .frame(maxWidth: .infinity)
    }

    private var noResults: some View {
        VStack(spacing: 12) {
            Spacer()

            Image(systemName: "doc.text.magnifyingglass")
                .font(.system(size: 40, weight: .thin))
                .foregroundStyle(Theme.textSecondary.opacity(0.4))

            Text("No results for \"\(state.searchQuery)\"")
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            Text("Try different keywords or check your spelling")
                .font(.system(size: 12))
                .foregroundStyle(Theme.textSecondary.opacity(0.7))

            Spacer()
        }
        .frame(maxWidth: .infinity)
    }
}
