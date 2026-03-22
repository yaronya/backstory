// swift-tools-version: 5.4
import PackageDescription

let package = Package(
    name: "Backstory",
    platforms: [.macOS(.v14)],
    targets: [
        .executableTarget(
            name: "Backstory",
            path: "Sources/Backstory"
        )
    ]
)
