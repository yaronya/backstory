import Foundation

struct ConfigService {
    static func detectRepoPath() -> String? {
        let homeDir = FileManager.default.homeDirectoryForCurrentUser.path
        let candidates = [
            ProcessInfo.processInfo.environment["BACKSTORY_REPO"],
            (homeDir as NSString).appendingPathComponent(".backstory/repo"),
            (homeDir as NSString).appendingPathComponent("backstory-decisions")
        ]

        for candidate in candidates {
            guard let path = candidate else { continue }
            if FileManager.default.fileExists(atPath: path) {
                return path
            }
        }

        return nil
    }

    static func detectBinaryPath() -> String? {
        let candidates = [
            "/usr/local/bin/backstory",
            "/opt/homebrew/bin/backstory",
            (FileManager.default.homeDirectoryForCurrentUser.path as NSString)
                .appendingPathComponent("go/bin/backstory")
        ]

        for path in candidates {
            if FileManager.default.fileExists(atPath: path) {
                return path
            }
        }

        return nil
    }
}
