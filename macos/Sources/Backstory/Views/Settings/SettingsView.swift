import SwiftUI

struct SettingsView: View {
    @Bindable var state: AppState
    @State private var syncOutput: String = ""
    @State private var isSyncing: Bool = false
    @State private var showApiKey: Bool = false

    var body: some View {
        VStack(spacing: 0) {
            header

            Divider().overlay(Theme.border)

            ScrollView {
                VStack(alignment: .leading, spacing: 28) {
                    repoSection
                    apiKeySection
                    syncSection
                    aboutSection
                }
                .padding(24)
            }
        }
        .background(Theme.background)
    }

    private var header: some View {
        HStack {
            Text("Settings")
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)
            Spacer()
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 14)
    }

    private var repoSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Repository Path")
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)

            Text("Path to your Backstory decisions repository")
                .font(.system(size: 12))
                .foregroundStyle(Theme.textSecondary)

            HStack(spacing: 8) {
                TextField("~/backstory-decisions", text: $state.repoPath)
                    .textFieldStyle(.plain)
                    .font(.system(size: 13))
                    .foregroundStyle(Theme.textPrimary)
                    .padding(10)
                    .background(
                        RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                            .fill(Theme.surfaceElevated)
                    )
                    .overlay(
                        RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                            .strokeBorder(Theme.border, lineWidth: 1)
                    )

                Button {
                    pickFolder()
                } label: {
                    Image(systemName: "folder")
                        .font(.system(size: 14))
                        .foregroundStyle(Theme.accent)
                        .padding(10)
                        .background(
                            RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                                .fill(Theme.surfaceElevated)
                        )
                        .overlay(
                            RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                                .strokeBorder(Theme.border, lineWidth: 1)
                        )
                }
                .buttonStyle(.plain)
            }
            .onChange(of: state.repoPath) {
                state.saveSettings()
            }
        }
    }

    private var apiKeySection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Claude API Key")
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)

            Text("Your Anthropic API key for the chat feature")
                .font(.system(size: 12))
                .foregroundStyle(Theme.textSecondary)

            HStack(spacing: 8) {
                Group {
                    if showApiKey {
                        TextField("sk-ant-...", text: $state.apiKey)
                    } else {
                        SecureField("sk-ant-...", text: $state.apiKey)
                    }
                }
                .textFieldStyle(.plain)
                .font(.system(size: 13, design: .monospaced))
                .foregroundStyle(Theme.textPrimary)
                .padding(10)
                .background(
                    RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                        .fill(Theme.surfaceElevated)
                )
                .overlay(
                    RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                        .strokeBorder(Theme.border, lineWidth: 1)
                )

                Button {
                    showApiKey.toggle()
                } label: {
                    Image(systemName: showApiKey ? "eye.slash" : "eye")
                        .font(.system(size: 14))
                        .foregroundStyle(Theme.textSecondary)
                        .padding(10)
                        .background(
                            RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                                .fill(Theme.surfaceElevated)
                        )
                        .overlay(
                            RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                                .strokeBorder(Theme.border, lineWidth: 1)
                        )
                }
                .buttonStyle(.plain)
            }
            .onChange(of: state.apiKey) {
                state.saveSettings()
            }
        }
    }

    private var syncSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Sync")
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)

            Text("Pull latest changes and reload decisions")
                .font(.system(size: 12))
                .foregroundStyle(Theme.textSecondary)

            HStack(spacing: 12) {
                Button {
                    performSync()
                } label: {
                    HStack(spacing: 6) {
                        if isSyncing {
                            ProgressView()
                                .controlSize(.small)
                        }
                        Text("Sync Now")
                            .font(.system(size: 13, weight: .medium))
                    }
                    .foregroundStyle(Color.white)
                    .padding(.horizontal, 20)
                    .padding(.vertical, 8)
                    .background(
                        RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                            .fill(isSyncing ? Theme.accent.opacity(0.5) : Theme.accent)
                    )
                }
                .buttonStyle(.plain)
                .disabled(isSyncing || state.repoPath.isEmpty)

                if let lastSynced = state.lastSynced {
                    Text("Last synced: \(lastSynced.formatted(date: .abbreviated, time: .shortened))")
                        .font(.system(size: 12))
                        .foregroundStyle(Theme.textSecondary)
                }
            }

            if !syncOutput.isEmpty {
                Text(syncOutput)
                    .font(.system(size: 11, design: .monospaced))
                    .foregroundStyle(Theme.textSecondary)
                    .padding(10)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(
                        RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                            .fill(Theme.surfaceElevated)
                    )
            }
        }
    }

    private var aboutSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("About")
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)

            HStack(spacing: 16) {
                VStack(alignment: .leading, spacing: 4) {
                    Text("Backstory")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(Theme.textPrimary)

                    Text("v1.0.0")
                        .font(.system(size: 12))
                        .foregroundStyle(Theme.textSecondary)
                }
            }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: Theme.cornerRadius)
                    .fill(Theme.surfaceElevated)
            )
        }
    }

    private func pickFolder() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.message = "Select your Backstory decisions repository"

        if panel.runModal() == .OK, let url = panel.url {
            state.repoPath = url.path
            state.saveSettings()
            reloadDecisions()
        }
    }

    private func performSync() {
        guard !state.repoPath.isEmpty else { return }
        isSyncing = true
        syncOutput = ""

        Task {
            do {
                let service = BackstoryService(repoPath: state.repoPath)
                let output = try await service.sync()
                let decisions = try await service.loadAllDecisions()

                await MainActor.run {
                    withAnimation(Theme.springAnimation) {
                        syncOutput = output.isEmpty ? "Sync complete" : output
                        state.decisions = decisions
                        state.lastSynced = Date()
                        isSyncing = false
                    }
                }
            } catch {
                await MainActor.run {
                    syncOutput = "Sync failed: \(error.localizedDescription)"
                    isSyncing = false
                }
            }
        }
    }

    private func reloadDecisions() {
        Task {
            let service = BackstoryService(repoPath: state.repoPath)
            do {
                let decisions = try await service.loadAllDecisions()
                await MainActor.run {
                    withAnimation(Theme.springAnimation) {
                        state.decisions = decisions
                    }
                }
            } catch {
                // silently handle
            }
        }
    }
}
