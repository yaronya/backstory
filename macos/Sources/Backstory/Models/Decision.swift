import Foundation

struct Decision: Identifiable, Codable, Hashable {
    var id: String { filePath }
    let type: String
    let date: String
    let author: String
    let anchor: String
    let linearIssue: String?
    let stale: Bool
    let title: String
    let body: String
    let filePath: String

    var typeLabel: String {
        type.capitalized
    }

    var parsedDate: Date? {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.date(from: date)
    }

    var formattedDate: String {
        guard let parsed = parsedDate else { return date }
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        return formatter.string(from: parsed)
    }
}
