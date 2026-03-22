import SwiftUI

struct MessageBubble: View {
    let message: ChatMessage
    @State private var appeared = false

    private var isUser: Bool { message.role == .user }

    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            if isUser { Spacer(minLength: 60) }

            VStack(alignment: isUser ? .trailing : .leading, spacing: 4) {
                Text(parseMarkdown(message.content))
                    .font(.system(size: 13))
                    .foregroundStyle(isUser ? Color.white : Theme.textPrimary)
                    .textSelection(.enabled)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 10)
                    .background(
                        RoundedRectangle(cornerRadius: 16)
                            .fill(isUser ? Theme.accent : Theme.surfaceElevated)
                    )

                Text(message.formattedTime)
                    .font(.system(size: 10))
                    .foregroundStyle(Theme.textSecondary.opacity(0.6))
                    .padding(.horizontal, 4)
            }

            if !isUser { Spacer(minLength: 60) }
        }
        .opacity(appeared ? 1 : 0)
        .offset(y: appeared ? 0 : 10)
        .onAppear {
            withAnimation(Theme.springAnimation) {
                appeared = true
            }
        }
    }

    private func parseMarkdown(_ text: String) -> AttributedString {
        do {
            return try AttributedString(markdown: text, options: .init(interpretedSyntax: .inlineOnlyPreservingWhitespace))
        } catch {
            return AttributedString(text)
        }
    }
}
