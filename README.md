<p align="center">
  <h1 align="center">Backstory</h1>
  <p align="center">Shared team memory for AI coding agents</p>
  <p align="center">
    <a href="https://github.com/yaronya/backstory/actions"><img src="https://github.com/yaronya/backstory/workflows/CI/badge.svg" alt="CI"></a>
    <a href="https://goreportcard.com/report/github.com/yaronya/backstory"><img src="https://goreportcard.com/badge/github.com/yaronya/backstory" alt="Go Report Card"></a>
    <a href="https://github.com/yaronya/backstory/blob/master/LICENSE"><img src="https://img.shields.io/github/license/yaronya/backstory" alt="License"></a>
    <a href="https://github.com/yaronya/backstory/releases"><img src="https://img.shields.io/github/v/release/yaronya/backstory" alt="Release"></a>
    <img src="https://img.shields.io/badge/go-%3E%3D1.26-blue" alt="Go Version">
  </p>
</p>

---

Your codebase captures **what** was built. Backstory captures **why**.

Backstory automatically extracts the reasoning behind code decisions during AI coding sessions and shares them across your team. When a teammate's agent starts a session, it already knows what you decided — and why.

**The Waze model:** you connect to get better routes (your agent gets smarter). You contribute passively by coding (decisions are captured automatically). Nobody "shares" anything — it just happens.

## How It Works

```
  Developer A's session                    Developer B's session
  ┌─────────────────────┐                  ┌─────────────────────┐
  │ Claude Code          │                  │ Claude Code          │
  │                      │                  │                      │
  │ "I chose SQS over    │   ┌──────────┐  │ Agent already knows: │
  │  direct invocation   │──▶│ Backstory │──│ "SQS was chosen for  │
  │  because of rate     │   │ Decisions │  │  rate limit reasons" │
  │  limits at 100 rps"  │   │   Repo    │  │                      │
  └─────────────────────┘   └──────────┘  └─────────────────────┘
                                 ▲
                            ┌────┴────┐
                            │ PM adds │
                            │ product │
                            │ context │
                            └─────────┘
```

1. **Session starts** — Backstory pulls the latest decisions and injects relevant context into your agent
2. **You code** — Your agent makes decisions with full team context
3. **Session ends** — Backstory extracts decisions and asks you to confirm sharing
4. **Teammate starts** — Their agent already knows what you decided

## Features

- **Auto-capture** — Extracts decisions from coding sessions via Claude API
- **Code-anchored** — Decisions are linked to specific directories, files, or services
- **Full-text search** — SQLite FTS5 index for fast keyword search
- **Linear integration** — Pulls issue context when working on a ticket
- **Pending queue** — Offline-resilient; queues decisions when network is unavailable
- **Staleness detection** — Flags decisions when their anchored code changes
- **Git-native** — Decisions repo is just a git repo with markdown files

## Install

**From source:**

```bash
go install github.com/yaronya/backstory/cmd/backstory@latest
```

**Clone and build:**

```bash
git clone https://github.com/yaronya/backstory.git
cd backstory
make build
```

**Requirements:** Go 1.26+, Git

## Quick Start

```bash
# 1. Create a new decisions repo
backstory init --path my-team-decisions

# 2. Or join an existing team's repo
backstory init --connect git@github.com:my-team/decisions.git

# 3. Set the environment variable (add to your shell profile)
export BACKSTORY_REPO=/path/to/my-team-decisions

# 4. Add your API keys
cat > /path/to/my-team-decisions/.backstory/config.local.yml << 'EOF'
claude_api_key: sk-ant-...
linear_api_key: lin_api_...
EOF

# 5. Build the search index
backstory index

# 6. Search decisions
backstory search "payment processing"
```

## Claude Code Integration

Add hooks to `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "backstory sync && backstory inject"
      }
    ],
    "Stop": [
      {
        "type": "command",
        "command": "backstory capture"
      }
    ]
  }
}
```

When a session starts, Backstory syncs the latest decisions and injects relevant context. When a session ends, it extracts decisions from the transcript and prompts you to share them.

## Commands

| Command | Description |
|---------|-------------|
| `backstory init` | Create a new decisions repo or connect to an existing one |
| `backstory sync` | Pull latest, process pending decisions, rebuild index |
| `backstory index` | Rebuild the local FTS5 search index |
| `backstory search <query>` | Search decisions by keyword |
| `backstory inject` | Output relevant decisions as XML context (used by hooks) |
| `backstory capture` | Extract decisions from a session transcript (used by hooks) |
| `backstory add` | Manually add a decision |
| `backstory status` | Show repo status, decision counts, and pending items |
| `backstory stale` | Detect and mark stale decisions |
| `backstory edit <file>` | Edit a decision file and commit the change |

## Configuration

**Team settings** (`.backstory/config.yml` — committed):

```yaml
team: my-team
repos:
  - name: backend
    url: git@github.com:my-team/backend.git
  - name: frontend
    url: git@github.com:my-team/frontend.git
linear:
  team_key: ENG
inject:
  max_decisions: 10
  max_tokens: 2000
extract:
  model: claude-haiku-4-5-20251001
  max_tokens: 4096
```

**Local secrets** (`.backstory/config.local.yml` — gitignored):

```yaml
claude_api_key: sk-ant-...
linear_api_key: lin_api_...
```

## Decisions Repo Structure

```
my-team-decisions/
├── product/
│   └── payments/
│       ├── 2026-03-20-stripe-over-adyen.md
│       └── 2026-03-22-no-bulk-ops-v1.md
├── technical/
│   └── backend/services/payment-service/
│       ├── 2026-03-19-sqs-over-direct-invocation.md
│       └── 2026-03-22-exponential-backoff-webhooks.md
├── .backstory/
│   ├── config.yml
│   ├── config.local.yml  (gitignored)
│   └── index.db           (gitignored)
└── README.md
```

Each decision file is markdown with YAML frontmatter:

```markdown
---
type: technical
date: 2026-03-19
author: sarah
anchor: backend/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation for vendor API

The vendor API rate-limits at 100 req/s. Direct invocation from the Lambda
would hit this limit during peak hours. SQS provides natural backpressure
and retry semantics without custom rate-limiting code.

Alternatives considered: Direct invocation with client-side rate limiting
— rejected because it adds complexity and doesn't handle concurrency spikes.
```

## Architecture

```
CLI (Go binary)
├── SessionStart hook → sync + inject relevant decisions
├── Stop hook         → capture decisions from session
├── search            → FTS5 full-text search
├── add               → manual decision entry
└── stale             → staleness detection

Storage
├── Git repo          → decisions as markdown (source of truth)
└── SQLite + FTS5     → local search index (disposable cache)

Integrations
├── Claude API        → decision extraction from transcripts
└── Linear API        → pull issue context on demand
```

## Contributing

Contributions are welcome. Please open an issue to discuss changes before submitting a PR.

```bash
# Run tests
make test

# Build
make build

# Run
./bin/backstory --help
```

## License

[MIT](LICENSE)
