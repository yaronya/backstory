import SwiftUI

struct ChatView: View {
    @Bindable var state: AppState

    var body: some View {
        VStack(spacing: 0) {
            chatHeader

            Divider().overlay(Theme.border)

            if state.chatMessages.isEmpty {
                welcomeState
            } else {
                messageList
            }

            Divider().overlay(Theme.border)

            ChatInput(state: state, onSend: sendMessage)
        }
        .background(Theme.background)
    }

    private var chatHeader: some View {
        HStack {
            Text("Chat")
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(Theme.textPrimary)

            Spacer()

            if !state.chatMessages.isEmpty {
                Button {
                    withAnimation(Theme.springAnimation) {
                        state.chatMessages.removeAll()
                    }
                } label: {
                    Image(systemName: "trash")
                        .font(.system(size: 13))
                        .foregroundStyle(Theme.textSecondary)
                }
                .buttonStyle(.plain)
                .help("Clear chat")
            }
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 14)
    }

    private var welcomeState: some View {
        VStack(spacing: 16) {
            Spacer()

            Image(systemName: "bubble.left.and.bubble.right")
                .font(.system(size: 48, weight: .thin))
                .foregroundStyle(Theme.accent.opacity(0.6))

            Text("Ask me anything about your team's decisions")
                .font(.system(size: 16, weight: .medium))
                .foregroundStyle(Theme.textSecondary)

            Text("I can search decisions, summarize context, and help you add new ones.")
                .font(.system(size: 13))
                .foregroundStyle(Theme.textSecondary.opacity(0.7))
                .multilineTextAlignment(.center)
                .frame(maxWidth: 400)

            Spacer()
        }
        .frame(maxWidth: .infinity)
    }

    private var messageList: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(spacing: 12) {
                    ForEach(state.chatMessages) { message in
                        MessageBubble(message: message)
                            .id(message.id)
                    }

                    if state.isTyping {
                        TypingIndicator()
                            .id("typing")
                    }
                }
                .padding(20)
            }
            .onChange(of: state.chatMessages.count) {
                withAnimation(Theme.springAnimation) {
                    if state.isTyping {
                        proxy.scrollTo("typing", anchor: .bottom)
                    } else if let last = state.chatMessages.last {
                        proxy.scrollTo(last.id, anchor: .bottom)
                    }
                }
            }
            .onChange(of: state.isTyping) {
                withAnimation(Theme.springAnimation) {
                    if state.isTyping {
                        proxy.scrollTo("typing", anchor: .bottom)
                    }
                }
            }
        }
    }

    private func sendMessage(_ text: String) {
        let userMessage = ChatMessage(role: .user, content: text)
        withAnimation(Theme.springAnimation) {
            state.chatMessages.append(userMessage)
            state.isTyping = true
        }

        Task {
            do {
                var context = ""
                if !state.repoPath.isEmpty {
                    let service = BackstoryService(repoPath: state.repoPath)
                    let relevant = try await service.search(query: text)
                    context = relevant.prefix(5).map { d in
                        "## \(d.title)\nType: \(d.type) | Author: \(d.author) | Date: \(d.date) | Anchor: \(d.anchor)\n\n\(d.body)"
                    }.joined(separator: "\n\n---\n\n")
                }

                guard !state.apiKey.isEmpty else {
                    await MainActor.run {
                        withAnimation(Theme.springAnimation) {
                            state.isTyping = false
                            state.chatMessages.append(
                                ChatMessage(role: .assistant, content: "Please configure your Claude API key in Settings to use chat.")
                            )
                        }
                    }
                    return
                }

                let claude = ClaudeService(apiKey: state.apiKey)
                let response = try await claude.chat(messages: state.chatMessages, context: context)

                await MainActor.run {
                    withAnimation(Theme.springAnimation) {
                        state.isTyping = false
                        state.chatMessages.append(
                            ChatMessage(role: .assistant, content: response)
                        )
                    }
                }
            } catch {
                await MainActor.run {
                    withAnimation(Theme.springAnimation) {
                        state.isTyping = false
                        state.chatMessages.append(
                            ChatMessage(role: .assistant, content: "Error: \(error.localizedDescription)")
                        )
                    }
                }
            }
        }
    }
}

struct TypingIndicator: View {
    @State private var dotIndex = 0
    private let timer = Timer.publish(every: 0.4, on: .main, in: .common).autoconnect()

    var body: some View {
        HStack {
            HStack(spacing: 4) {
                ForEach(0..<3, id: \.self) { index in
                    Circle()
                        .fill(Theme.textSecondary)
                        .frame(width: 6, height: 6)
                        .opacity(dotIndex == index ? 1.0 : 0.3)
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(
                RoundedRectangle(cornerRadius: 16)
                    .fill(Theme.surfaceElevated)
            )

            Spacer()
        }
        .onReceive(timer) { _ in
            withAnimation(.easeInOut(duration: 0.2)) {
                dotIndex = (dotIndex + 1) % 3
            }
        }
    }
}
