import SwiftUI

struct DecisionDetailView: View {
    let decision: Decision
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 0) {
            header

            Divider().overlay(Theme.border)

            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    metadataSection
                    bodySection
                }
                .padding(24)
            }
        }
        .frame(minWidth: 600, minHeight: 500)
        .background(Theme.background)
    }

    private var header: some View {
        HStack {
            Text(decision.title)
                .font(.system(size: 18, weight: .bold))
                .foregroundStyle(Theme.textPrimary)
                .lineLimit(2)

            Spacer()

            Button {
                NSWorkspace.shared.selectFile(decision.filePath, inFileViewerRootedAtPath: "")
            } label: {
                Image(systemName: "doc.on.doc")
                    .font(.system(size: 13))
                    .foregroundStyle(Theme.textSecondary)
            }
            .buttonStyle(.plain)
            .help("Reveal in Finder")

            Button {
                dismiss()
            } label: {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 18))
                    .foregroundStyle(Theme.textSecondary)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 16)
    }

    private var metadataSection: some View {
        HStack(spacing: 12) {
            TypeBadge(type: decision.type)

            if decision.stale {
                StaleBadge()
            }

            MetadataItem(icon: "calendar", text: decision.formattedDate)
            MetadataItem(icon: "person", text: decision.author)
            MetadataItem(icon: "link", text: decision.anchor)

            if let issue = decision.linearIssue, !issue.isEmpty {
                MetadataItem(icon: "tag", text: issue)
            }

            Spacer()
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: Theme.cornerRadius)
                .fill(Theme.surfaceElevated)
        )
    }

    private var bodySection: some View {
        Text(parseMarkdown(decision.body))
            .font(.system(size: 14))
            .foregroundStyle(Theme.textPrimary)
            .textSelection(.enabled)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func parseMarkdown(_ text: String) -> AttributedString {
        do {
            var result = try AttributedString(markdown: text, options: .init(interpretedSyntax: .inlineOnlyPreservingWhitespace))
            result.foregroundColor = NSColor(Theme.textPrimary)
            return result
        } catch {
            return AttributedString(text)
        }
    }
}

struct MetadataItem: View {
    let icon: String
    let text: String

    var body: some View {
        Label(text, systemImage: icon)
            .font(.system(size: 12))
            .foregroundStyle(Theme.textSecondary)
    }
}
