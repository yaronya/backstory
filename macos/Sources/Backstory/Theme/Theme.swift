import SwiftUI

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 6:
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8:
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (255, 0, 0, 0)
        }
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}

enum Theme {
    static let background = Color(hex: "020617")
    static let surface = Color(hex: "0F172A")
    static let surfaceElevated = Color(hex: "1E293B")
    static let border = Color(hex: "334155")
    static let textPrimary = Color(hex: "F8FAFC")
    static let textSecondary = Color(hex: "94A3B8")
    static let accent = Color(hex: "8B5CF6")
    static let accentSecondary = Color(hex: "E94560")
    static let success = Color(hex: "22C55E")
    static let warning = Color(hex: "F59E0B")

    static let sidebarWidth: CGFloat = 220
    static let cornerRadius: CGFloat = 10
    static let cornerRadiusSmall: CGFloat = 6
    static let animationDuration: Double = 0.3

    static let springAnimation = Animation.spring(duration: 0.3)
}
