import SwiftUI

struct SidebarView: View {
    @Bindable var state: AppState

    var body: some View {
        VStack(spacing: 0) {
            headerSection

            Divider()
                .overlay(Theme.border)
                .padding(.horizontal, 16)

            navigationSection

            Spacer()

            syncSection
        }
        .background(Theme.surface)
    }

    private var headerSection: some View {
        HStack(spacing: 10) {
            Image(systemName: "book.closed.fill")
                .font(.system(size: 20, weight: .semibold))
                .foregroundStyle(Theme.accent)

            Text("Backstory")
                .font(.system(size: 18, weight: .bold))
                .foregroundStyle(Theme.textPrimary)

            Spacer()
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 20)
    }

    private var navigationSection: some View {
        VStack(spacing: 4) {
            ForEach(Tab.allCases) { tab in
                SidebarButton(
                    tab: tab,
                    isSelected: state.selectedTab == tab
                ) {
                    withAnimation(Theme.springAnimation) {
                        state.selectedTab = tab
                    }
                }
            }
        }
        .padding(.horizontal, 12)
        .padding(.top, 12)
    }

    private var syncSection: some View {
        VStack(spacing: 6) {
            Divider()
                .overlay(Theme.border)
                .padding(.horizontal, 16)

            HStack(spacing: 6) {
                Circle()
                    .fill(state.repoPath.isEmpty ? Theme.warning : Theme.success)
                    .frame(width: 6, height: 6)

                Text(syncStatusText)
                    .font(.system(size: 11))
                    .foregroundStyle(Theme.textSecondary)

                Spacer()
            }
            .padding(.horizontal, 16)
            .padding(.bottom, 16)
            .padding(.top, 8)
        }
    }

    private var syncStatusText: String {
        if state.repoPath.isEmpty {
            return "No repo configured"
        }
        guard let lastSynced = state.lastSynced else {
            return "Not synced yet"
        }
        let minutes = Int(Date().timeIntervalSince(lastSynced) / 60)
        if minutes < 1 {
            return "Synced just now"
        }
        return "Synced \(minutes)m ago"
    }
}

struct SidebarButton: View {
    let tab: Tab
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 10) {
                Image(systemName: tab.icon)
                    .font(.system(size: 14, weight: .medium))
                    .frame(width: 20)

                Text(tab.label)
                    .font(.system(size: 13, weight: isSelected ? .semibold : .regular))

                Spacer()
            }
            .foregroundStyle(isSelected ? Theme.textPrimary : Theme.textSecondary)
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .background(
                RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                    .fill(isSelected ? Theme.accent.opacity(0.2) : Color.clear)
            )
            .overlay(
                RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                    .strokeBorder(isSelected ? Theme.accent.opacity(0.3) : Color.clear, lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
    }
}
