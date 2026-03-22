import SwiftUI

struct DecisionRow: View {
    let decision: Decision

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 8) {
                TypeBadge(type: decision.type)

                if decision.stale {
                    StaleBadge()
                }

                Spacer()

                Text(decision.formattedDate)
                    .font(.system(size: 11))
                    .foregroundStyle(Theme.textSecondary)
            }

            Text(decision.title)
                .font(.system(size: 13, weight: .medium))
                .foregroundStyle(Theme.textPrimary)
                .lineLimit(2)

            HStack(spacing: 12) {
                Label(decision.author, systemImage: "person")
                Label(decision.anchor, systemImage: "link")
            }
            .font(.system(size: 11))
            .foregroundStyle(Theme.textSecondary)
            .lineLimit(1)
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: Theme.cornerRadius)
                .fill(Theme.surfaceElevated)
        )
        .overlay(
            RoundedRectangle(cornerRadius: Theme.cornerRadius)
                .strokeBorder(Theme.border.opacity(0.5), lineWidth: 1)
        )
    }
}

struct TypeBadge: View {
    let type: String

    var body: some View {
        Text(type.capitalized)
            .font(.system(size: 10, weight: .semibold))
            .foregroundStyle(badgeColor)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(
                Capsule().fill(badgeColor.opacity(0.15))
            )
    }

    private var badgeColor: Color {
        type.lowercased() == "product" ? Theme.accent : Theme.accentSecondary
    }
}

struct StaleBadge: View {
    var body: some View {
        Text("Stale")
            .font(.system(size: 10, weight: .semibold))
            .foregroundStyle(Theme.warning)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(
                Capsule().fill(Theme.warning.opacity(0.15))
            )
    }
}
