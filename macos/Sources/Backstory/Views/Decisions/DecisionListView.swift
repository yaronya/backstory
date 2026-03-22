import SwiftUI

struct DecisionListView: View {
    @Bindable var state: AppState

    var body: some View {
        VStack(spacing: 0) {
            header

            Divider().overlay(Theme.border)

            filterBar

            if state.filteredDecisions.isEmpty {
                emptyState
            } else {
                decisionsList
            }
        }
        .background(Theme.background)
        .sheet(isPresented: $state.showAddDecision) {
            AddDecisionView(state: state)
        }
        .sheet(item: $state.selectedDecision) { decision in
            DecisionDetailView(decision: decision)
        }
    }

    private var header: some View {
        HStack {
            Text("Decisions")
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)

            Text("\(state.filteredDecisions.count)")
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(Theme.textSecondary)
                .padding(.horizontal, 8)
                .padding(.vertical, 2)
                .background(Capsule().fill(Theme.surfaceElevated))

            Spacer()

            Button {
                withAnimation(Theme.springAnimation) {
                    state.showAddDecision = true
                }
            } label: {
                Image(systemName: "plus.circle.fill")
                    .font(.system(size: 18))
                    .foregroundStyle(Theme.accent)
            }
            .buttonStyle(.plain)
            .help("Add decision")
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 14)
    }

    private var filterBar: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                ForEach(AppState.DecisionFilter.allCases, id: \.self) { filter in
                    FilterPill(
                        label: filter.rawValue,
                        isSelected: state.decisionFilter == filter
                    ) {
                        withAnimation(Theme.springAnimation) {
                            state.decisionFilter = filter
                        }
                    }
                }
            }
            .padding(.horizontal, 20)
            .padding(.vertical, 10)
        }
    }

    private var decisionsList: some View {
        ScrollView {
            LazyVStack(spacing: 8) {
                ForEach(state.filteredDecisions) { decision in
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
    }

    private var emptyState: some View {
        VStack(spacing: 12) {
            Spacer()

            Image(systemName: "doc.text.magnifyingglass")
                .font(.system(size: 40, weight: .thin))
                .foregroundStyle(Theme.textSecondary.opacity(0.4))

            Text("No decisions found")
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            if state.repoPath.isEmpty {
                Text("Configure your repo path in Settings to load decisions.")
                    .font(.system(size: 12))
                    .foregroundStyle(Theme.textSecondary.opacity(0.7))
            }

            Spacer()
        }
        .frame(maxWidth: .infinity)
    }
}

struct FilterPill: View {
    let label: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(label)
                .font(.system(size: 12, weight: isSelected ? .semibold : .regular))
                .foregroundStyle(isSelected ? Theme.textPrimary : Theme.textSecondary)
                .padding(.horizontal, 14)
                .padding(.vertical, 6)
                .background(
                    Capsule().fill(isSelected ? Theme.accent.opacity(0.2) : Theme.surfaceElevated)
                )
                .overlay(
                    Capsule().strokeBorder(isSelected ? Theme.accent.opacity(0.4) : Theme.border.opacity(0.5), lineWidth: 1)
                )
        }
        .buttonStyle(.plain)
    }
}
