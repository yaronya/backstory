import SwiftUI

struct AddDecisionView: View {
    @Bindable var state: AppState
    @Environment(\.dismiss) private var dismiss
    @State private var type: String = "product"
    @State private var title: String = ""
    @State private var bodyText: String = ""
    @State private var anchor: String = ""
    @State private var linearIssue: String = ""
    @State private var isSaving: Bool = false
    @State private var errorMessage: String? = nil

    var body: some View {
        VStack(spacing: 0) {
            header

            Divider().overlay(Theme.border)

            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    typePicker
                    titleField
                    anchorField
                    linearIssueField
                    bodyField

                    if let error = errorMessage {
                        Text(error)
                            .font(.system(size: 12))
                            .foregroundStyle(Theme.accentSecondary)
                    }
                }
                .padding(24)
            }

            Divider().overlay(Theme.border)

            footer
        }
        .frame(minWidth: 500, minHeight: 550)
        .background(Theme.background)
    }

    private var header: some View {
        HStack {
            Text("Add Decision")
                .font(.system(size: 18, weight: .bold))
                .foregroundStyle(Theme.textPrimary)

            Spacer()

            Button { dismiss() } label: {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 18))
                    .foregroundStyle(Theme.textSecondary)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 16)
    }

    private var typePicker: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Type")
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            Picker("", selection: $type) {
                Text("Product").tag("product")
                Text("Technical").tag("technical")
            }
            .pickerStyle(.segmented)
        }
    }

    private var titleField: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Title")
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            TextField("Decision title", text: $title)
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
        }
    }

    private var anchorField: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Anchor")
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            TextField("e.g., team-name, project-name", text: $anchor)
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
        }
    }

    private var linearIssueField: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Linear Issue (optional)")
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            TextField("e.g., ENG-123", text: $linearIssue)
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
        }
    }

    private var bodyField: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Body")
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            TextEditor(text: $bodyText)
                .font(.system(size: 13))
                .foregroundStyle(Theme.textPrimary)
                .scrollContentBackground(.hidden)
                .frame(minHeight: 150)
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
    }

    private var footer: some View {
        HStack {
            Spacer()

            Button("Cancel") { dismiss() }
                .buttonStyle(.plain)
                .foregroundStyle(Theme.textSecondary)

            Button {
                saveDecision()
            } label: {
                HStack(spacing: 6) {
                    if isSaving {
                        ProgressView()
                            .controlSize(.small)
                    }
                    Text("Add Decision")
                        .font(.system(size: 13, weight: .semibold))
                }
                .foregroundStyle(Color.white)
                .padding(.horizontal, 20)
                .padding(.vertical, 8)
                .background(
                    RoundedRectangle(cornerRadius: Theme.cornerRadiusSmall)
                        .fill(canSave ? Theme.accent : Theme.accent.opacity(0.3))
                )
            }
            .buttonStyle(.plain)
            .disabled(!canSave)
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 14)
    }

    private var canSave: Bool {
        !title.isEmpty && !anchor.isEmpty && !bodyText.isEmpty && !isSaving
    }

    private func saveDecision() {
        guard !state.repoPath.isEmpty else {
            errorMessage = "Configure your repo path in Settings first."
            return
        }

        isSaving = true
        errorMessage = nil

        Task {
            do {
                let service = BackstoryService(repoPath: state.repoPath)
                try await service.addDecision(
                    type: type,
                    title: title,
                    body: bodyText,
                    anchor: anchor,
                    linearIssue: linearIssue.isEmpty ? nil : linearIssue
                )
                let updated = try await service.loadAllDecisions()

                await MainActor.run {
                    withAnimation(Theme.springAnimation) {
                        state.decisions = updated
                        isSaving = false
                    }
                    dismiss()
                }
            } catch {
                await MainActor.run {
                    errorMessage = "Failed to save: \(error.localizedDescription)"
                    isSaving = false
                }
            }
        }
    }
}
