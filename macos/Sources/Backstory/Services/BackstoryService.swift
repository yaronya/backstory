import Foundation

actor BackstoryService {
    let repoPath: String

    init(repoPath: String) {
        self.repoPath = repoPath
    }

    func loadAllDecisions() async throws -> [Decision] {
        let decisionsPath = (repoPath as NSString).appendingPathComponent("decisions")
        let fileManager = FileManager.default
        guard fileManager.fileExists(atPath: decisionsPath) else {
            return []
        }

        var results: [Decision] = []
        let enumerator = fileManager.enumerator(atPath: decisionsPath)

        while let element = enumerator?.nextObject() as? String {
            guard element.hasSuffix(".md") else { continue }
            let fullPath = (decisionsPath as NSString).appendingPathComponent(element)
            if let decision = try? parseDecisionFile(at: fullPath) {
                results.append(decision)
            }
        }

        return results.sorted { ($0.parsedDate ?? .distantPast) > ($1.parsedDate ?? .distantPast) }
    }

    func search(query: String) async throws -> [Decision] {
        let all = try await loadAllDecisions()
        guard !query.isEmpty else { return all }
        let q = query.lowercased()
        return all.filter { d in
            d.title.lowercased().contains(q) ||
            d.body.lowercased().contains(q) ||
            d.author.lowercased().contains(q) ||
            d.anchor.lowercased().contains(q)
        }
    }

    func addDecision(type: String, title: String, body: String, anchor: String, linearIssue: String?) async throws {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        let dateStr = dateFormatter.string(from: Date())

        let slug = title.lowercased()
            .replacingOccurrences(of: " ", with: "-")
            .filter { $0.isLetter || $0.isNumber || $0 == "-" }

        let filename = "\(dateStr)-\(slug).md"
        let dirPath = (repoPath as NSString).appendingPathComponent("decisions/\(type)")

        let fileManager = FileManager.default
        if !fileManager.fileExists(atPath: dirPath) {
            try fileManager.createDirectory(atPath: dirPath, withIntermediateDirectories: true)
        }

        let filePath = (dirPath as NSString).appendingPathComponent(filename)

        var content = "---\n"
        content += "type: \(type)\n"
        content += "date: \(dateStr)\n"
        content += "author: \"\"\n"
        content += "anchor: \(anchor)\n"
        if let issue = linearIssue, !issue.isEmpty {
            content += "linear_issue: \(issue)\n"
        }
        content += "stale: false\n"
        content += "title: \(title)\n"
        content += "---\n\n"
        content += body

        try content.write(toFile: filePath, atomically: true, encoding: .utf8)
    }

    func sync() async throws -> String {
        return try await runCommand("git", arguments: ["-C", repoPath, "pull", "--rebase"])
    }

    func status() async throws -> String {
        return try await runCommand("git", arguments: ["-C", repoPath, "status", "--short"])
    }

    private func parseDecisionFile(at path: String) throws -> Decision {
        let content = try String(contentsOfFile: path, encoding: .utf8)
        let parts = content.components(separatedBy: "---")

        guard parts.count >= 3 else {
            throw ServiceError.invalidFormat
        }

        let frontmatter = parts[1]
        let body = parts.dropFirst(2).joined(separator: "---").trimmingCharacters(in: .whitespacesAndNewlines)

        var metadata: [String: String] = [:]
        for line in frontmatter.components(separatedBy: "\n") {
            let trimmed = line.trimmingCharacters(in: .whitespaces)
            guard let colonIndex = trimmed.firstIndex(of: ":") else { continue }
            let key = String(trimmed[trimmed.startIndex..<colonIndex]).trimmingCharacters(in: .whitespaces)
            let value = String(trimmed[trimmed.index(after: colonIndex)...])
                .trimmingCharacters(in: .whitespaces)
                .trimmingCharacters(in: CharacterSet(charactersIn: "\""))
            metadata[key] = value
        }

        return Decision(
            type: metadata["type"] ?? "technical",
            date: metadata["date"] ?? "",
            author: metadata["author"] ?? "",
            anchor: metadata["anchor"] ?? "",
            linearIssue: metadata["linear_issue"],
            stale: metadata["stale"]?.lowercased() == "true",
            title: metadata["title"] ?? (path as NSString).lastPathComponent,
            body: body,
            filePath: path
        )
    }

    private func runCommand(_ command: String, arguments: [String]) async throws -> String {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        process.arguments = [command] + arguments

        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = pipe

        try process.run()
        process.waitUntilExit()

        let data = pipe.fileHandleForReading.readDataToEndOfFile()
        return String(data: data, encoding: .utf8) ?? ""
    }

    enum ServiceError: Error {
        case invalidFormat
        case commandFailed(String)
    }
}
