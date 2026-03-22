import Foundation

actor ClaudeService {
    let apiKey: String
    private let baseURL = "https://api.anthropic.com/v1/messages"
    private let model = "claude-sonnet-4-20250514"

    init(apiKey: String) {
        self.apiKey = apiKey
    }

    func chat(messages: [ChatMessage], context: String) async throws -> String {
        let systemPrompt = """
        You are Backstory, an AI assistant for a team's decisions repository. \
        You help product managers and engineers understand, search, and add team decisions. \
        When asked about decisions, search the repo and provide answers with sources. \
        When asked to add a decision, confirm the details and add it.

        Here is context from the decisions repository:
        \(context)
        """

        let apiMessages = messages.map { msg -> [String: String] in
            ["role": msg.role.rawValue, "content": msg.content]
        }

        let requestBody: [String: Any] = [
            "model": model,
            "max_tokens": 4096,
            "system": systemPrompt,
            "messages": apiMessages
        ]

        let jsonData = try JSONSerialization.data(withJSONObject: requestBody)

        var request = URLRequest(url: URL(string: baseURL)!)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "content-type")
        request.setValue("2023-06-01", forHTTPHeaderField: "anthropic-version")
        request.setValue(apiKey, forHTTPHeaderField: "x-api-key")
        request.httpBody = jsonData

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            let errorText = String(data: data, encoding: .utf8) ?? "Unknown error"
            throw ClaudeError.apiError(errorText)
        }

        let json = try JSONSerialization.jsonObject(with: data) as? [String: Any]
        guard let contentArray = json?["content"] as? [[String: Any]],
              let firstContent = contentArray.first,
              let text = firstContent["text"] as? String else {
            throw ClaudeError.invalidResponse
        }

        return text
    }

    enum ClaudeError: Error, LocalizedError {
        case apiError(String)
        case invalidResponse

        var errorDescription: String? {
            switch self {
            case .apiError(let msg): return "API Error: \(msg)"
            case .invalidResponse: return "Invalid response from Claude API"
            }
        }
    }
}
