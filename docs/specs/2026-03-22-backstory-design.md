# Backstory — Design Spec

**Date:** 2026-03-22
**Status:** Draft

## Problem

Code repositories capture *what* was built but not *why*. The reasoning behind architectural decisions, product tradeoffs, and implementation choices lives in developers' heads, scattered Slack threads, and stale Linear tickets. When a new developer (or their AI coding agent) picks up a task, they lack the context that informed how the codebase got to its current state.

Single-user AI memory tools exist (claude-mem, Sugar, Engram) but they don't solve the team problem: one developer's discoveries stay invisible to everyone else.

## Solution

Backstory is a shared team memory for AI coding agents. It automatically captures the reasoning behind code decisions during AI coding sessions and makes that context available to every developer's agent on the team. Product managers interact through a macOS app powered by Claude, adding product context without touching code or markdown.

**Core metaphor: Waze.** You connect to get better routes (your agent gets smarter from the team's collective knowledge). You contribute passively by driving (your coding sessions automatically produce decisions). Nobody thinks of themselves as "sharing" — it just happens.

## Architecture Overview

```
Decisions Repo (git, separate from code repos)
├── product/          ← PM decisions (written via macOS app + Claude)
├── technical/        ← Dev decisions (auto-captured from sessions)
└── .backstory/       ← Config, index metadata

CLI (Go binary — fast, low memory)
├── SessionStart hook → pulls relevant decisions, injects as context
├── Stop hook         → extracts decisions from session, commits to repo
├── search            → queries local SQLite index
├── sync              → git pull/push
└── index             → rebuilds local SQLite from markdown files

macOS App
├── Claude-powered chat interface over the decisions repo
├── PMs read, ask questions, add product decisions via natural language
├── Claude handles all markdown structure — PMs never see raw files
└── Connects to same CLI core for git operations and search

Linear/Slack Integration (pull-based)
├── When dev starts work on ENG-1234, CLI fetches Linear issue + linked threads
└── Injects PM context alongside repo decisions
```

## Decisions Repo

A dedicated git repository, separate from code repos. This is a prerequisite for Backstory to work.

### Access control

The decisions repo is a standard git repo. Access is managed through the git hosting platform (GitHub/GitLab):
- All team members (devs + PMs) get write access
- No branch protection on main — direct pushes required for frictionless capture
- `backstory init` sets up the repo with appropriate team permissions via `gh` CLI
- Authentication uses the developer's existing git credentials (SSH keys or HTTPS tokens)
- The macOS app uses the same git credentials configured on the machine

### Why separate

- **Branch protection kills the flow.** Code repos typically block direct pushes to main. The decisions repo must allow direct pushes for frictionless capture.
- **Different lifecycle.** Code has releases, rollbacks, feature branches. Decisions are append-mostly and linear.
- **Different audiences.** PMs need access to the decisions repo without needing write access to code repos.
- **Multi-repo teams.** Many teams have multiple code repos but want one shared decisions memory.

### Structure

```
backstory-decisions/
├── product/
│   ├── payments/
│   │   ├── 2026-03-20-stripe-over-adyen.md
│   │   └── 2026-03-22-no-bulk-ops-v1.md
│   └── notifications/
│       └── 2026-03-18-email-only-for-launch.md
├── technical/
│   ├── env0/services/payment-service/
│   │   ├── 2026-03-19-sqs-over-direct-invocation.md
│   │   └── 2026-03-22-exponential-backoff-stripe-webhooks.md
│   └── env0/services/notification-service/
│       └── 2026-03-21-ses-template-approach.md
├── .backstory/
│   ├── config.yml        ← linked code repos, team settings (committed)
│   ├── config.local.yml  ← local overrides, API keys (gitignored)
│   └── index.db          ← local SQLite (gitignored)
└── README.md
```

### Decision file format

```markdown
---
type: technical | product
date: 2026-03-19
author: sarah
anchor: env0/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation for vendor API

The vendor API rate-limits at 100 req/s. Direct invocation from the Lambda
would hit this limit during peak hours. SQS provides natural backpressure
and retry semantics without custom rate-limiting code.

Considered alternatives:
- Direct invocation with client-side rate limiting — rejected because it
  adds complexity and doesn't handle Lambda concurrency spikes
- SNS + SQS fan-out — overkill for a single consumer
```

### Config schema

**`.backstory/config.yml`** (committed — shared team settings):
```yaml
team: my-team
repos:
  - name: env0
    url: git@github.com:env0/env0.git
  - name: frontend
    url: git@github.com:env0/frontend.git
linear:
  team_key: ENG
inject:
  max_decisions: 10
  max_tokens: 2000
staleness:
  archive_after_months: 6
  change_threshold: 0.5
```

**`.backstory/config.local.yml`** (gitignored — per-developer):
```yaml
claude_api_key: sk-ant-...
linear_api_key: lin_api_...
slack_token: xoxb-...          # optional, deferred to post-MVP
```

### Multi-repo mapping

The `repos` list in config maps repo names to URLs. When `backstory inject` runs:
1. Reads the current directory's git remote URL
2. Matches it against the `repos` list to determine the repo name
3. Queries the index for decisions anchored under that repo name (e.g., `env0/services/payment-service/`)

This allows a single decisions repo to serve multiple code repos. The anchor path format is always `<repo-name>/<path-within-repo>`.

### Context injection format

`backstory inject` outputs a structured block that Claude Code's SessionStart hook prepends to the session:

```
<backstory>
<decisions>
<decision type="technical" date="2026-03-19" author="sarah" anchor="env0/services/payment-service/">
Chose SQS over direct invocation for vendor API. The vendor API rate-limits at 100 req/s.
SQS provides natural backpressure and retry semantics.
</decision>
<decision type="product" date="2026-03-22" author="pm-david" anchor="payments" linear="ENG-892">
No bulk operations in v1. Too risky for initial launch. Single-item operations only.
</decision>
</decisions>
<linear issue="ENG-1234">
Title: Add Stripe webhook retry logic
Description: Implement exponential backoff for failed Stripe webhook deliveries...
</linear>
</backstory>
```

XML tags are used because Claude handles structured XML context well. The block is kept under the configured `max_tokens` limit (default 2000).

## CLI Tool

Written in Go. Single binary, fast startup (~5ms), low memory (~10MB).

### Commands

```
backstory init          ← initialize decisions repo from template
backstory sync          ← git pull/push
backstory search <query> ← semantic search over decisions
backstory index         ← rebuild SQLite index from markdown files
backstory inject        ← called by SessionStart hook, outputs relevant context
backstory capture       ← called by Stop hook, extracts and commits decisions
backstory status        ← show repo status, stale decisions, recent activity
```

### Claude Code Integration

Backstory integrates via hooks in Claude Code's `settings.json`:

**SessionStart hook:**
1. `backstory sync` — pulls latest from decisions repo
2. Checks pending queue for unconfirmed decisions from previous sessions
3. `backstory inject` — reads current code repo path, queries SQLite index for decisions anchored to relevant paths, outputs them as context injected into the session

**Stop hook:**
1. Receives session summary from Claude Code
2. Calls Claude API to extract candidate decisions (the hard AI problem)
3. Saves candidates to a local pending queue (`~/.backstory/pending/`)
4. Presents confirmation prompt in the terminal (blocking, 30-second timeout):

```
Backstory captured from this session:
  1. [x] Chose SQS over direct invocation for vendor API (rate limit 100 req/s)
  2. [x] Added exponential backoff for Stripe webhooks (can delay up to 5min)
  3. [ ] Fixed flaky test timing issue

[Share 1,2 to team] [Edit] [Dismiss all]
```

5. Confirmed items are written as markdown files and committed/pushed to the decisions repo
6. On timeout: items stay in pending queue, surfaced on next `backstory sync` or session start

**Default is "share"** — items come pre-checked. Developer only acts to remove.

**Failure modes:**
- Claude API unavailable: session transcript saved locally to `~/.backstory/pending/` for extraction on next sync
- Git push conflict: automatic `pull --rebase` and retry (up to 3 attempts). If still failing, items stay in pending queue
- No network: all operations queued locally, pushed on next `backstory sync`
- Terminal closed before confirmation: pending items surfaced on next session start

### Injection relevance algorithm

`backstory inject` must select the right decisions from potentially hundreds. The algorithm:

1. **Anchor match** — find decisions anchored to paths that are ancestors or descendants of the current working directory (e.g., working in `services/payment-service/handler.ts` matches anchors `services/payment-service/`, `services/payment-service/handler.ts`)
2. **Linear issue match** — if the current branch matches `eng-XXXX-*`, include decisions linked to that issue
3. **Recency boost** — newer decisions rank higher (exponential decay, half-life 30 days)
4. **Retrieval frequency** — decisions retrieved often by other team members rank higher (signals ongoing relevance)
5. **Exclude stale** — decisions with `stale: true` are excluded unless explicitly searched

Results are capped at `max_decisions` (default 10) and `max_tokens` (default 2000). If the relevant set exceeds the budget, lower-ranked decisions are dropped.

### Local SQLite Index

The SQLite database is a local, disposable cache built from the markdown files. It is gitignored.

**What it stores:**
- Full-text search index (SQLite FTS5) over decision content
- Embedding vectors for semantic search (sqlite-vec)
- Anchor paths for quick lookup by code location
- Staleness metadata

**Why SQLite:**
- Zero infrastructure — no server, no network dependency
- Fast — sub-millisecond queries for the SessionStart hook
- Disposable — `backstory index` rebuilds it from scratch anytime
- Each team member has their own local copy (slight inconsistency window after pushes, acceptable)

## Code Anchoring

Decisions are anchored to code locations. The anchor is a string path, not a filesystem dependency — it references paths in code repos without requiring the decisions repo to live alongside them.

### Granularity

Auto-detected by the extraction agent, defaulting to directory/service level:

| Decision type | Anchor level | Example |
|---|---|---|
| Architecture/design choice | Directory/service | `env0/services/payment-service/` |
| Implementation pattern | File | `env0/services/payment-service/handler.ts` |
| Specific logic reasoning | Function (fallback to file) | `env0/services/payment-service/handler.ts:processWebhook()` |
| Product decision | Feature area or Linear issue | `payments` or `ENG-892` |

### Staleness detection

- CLI watches code repos for significant changes to anchored paths (via git diff on sync)
- "Significant change" = file deleted, renamed, or >50% of lines changed (measured by git diff stat)
- When anchored code changes substantially, the decision's `stale` frontmatter flag is set to `true`
- Decisions not retrieved by any agent in 6 months are auto-archived (moved to `archive/` directory)
- Stale decisions are excluded from `inject` by default but still searchable

## macOS App

A native macOS application that provides a Claude-powered chat interface over the decisions repo.

### Target user

Product managers and non-technical team members who need to understand and contribute to the team's decision context without touching git, markdown, or the terminal.

### Interaction model

The app is a **chat with Claude about the decisions repo**, not a markdown editor.

**Example interactions:**

- *"What did the team decide about payment processing?"* → Claude searches decisions and answers with sources
- *"Add a product decision: we're going with Stripe for APAC because of multi-currency support. Related to ENG-892"* → Claude writes properly structured markdown, commits, pushes
- *"Why was SQS chosen for notifications?"* → Claude pulls both technical and product decisions
- *"What's changed in the payment service this week?"* → Claude summarizes recent technical decisions anchored to that area

**Claude handles all markdown structure.** PMs never see or edit raw files. This ensures consistent formatting and proper frontmatter.

### Technical approach

- Native macOS app (Swift/SwiftUI)
- Embeds or connects to the same Go CLI core for git operations and search
- Uses Claude API for the chat interface
- Local-first — the decisions repo is cloned locally

## Linear/Slack Integration

Pull-based. No webhooks, no background workers.

### How it works

When a developer starts working on a Linear ticket (detected via branch name pattern like `eng-1234-description`):

1. `backstory inject` detects the ticket reference
2. CLI calls Linear API to fetch the issue, its comments, and linked content
3. Optionally fetches linked Slack threads
4. Injects this PM context alongside repo decisions into the session

### Why pull-based

- Push-based creates a firehose — most Linear comments and Slack messages are noise
- Pull is precise — fetches only what's relevant at the moment it's needed
- No infrastructure overhead
- Privacy-friendly — not hoovering up all team communication

## Decision Types

Two types of entries, reflecting the two audiences:

### Product decisions
- Written by PMs through the macOS app
- Linked to Linear issues or feature areas
- Plain language: "We chose Stripe over Adyen for APAC because of multi-currency support"
- Anchored to feature areas, not code paths

### Technical decisions
- Auto-captured from developer coding sessions
- Linked to code locations (directories, files, functions)
- More structured: what was decided, what alternatives were considered, why
- Anchored to code paths with auto-detected granularity

Both types surface in the agent's context when relevant. A developer working on the payment service sees both "PM decided no bulk ops in v1 (ENG-892)" and "Sarah chose SQS here because of rate limits."

### Editing and retracting decisions

- Devs can edit/retract via `backstory edit <id>` or by editing the markdown file directly
- PMs can edit/retract via the macOS app chat: *"Update the Stripe decision — we've switched back to Adyen"*
- Edits are git commits — full history is preserved
- Retracted decisions are moved to `archive/` (not deleted) to preserve audit trail

## Target Users

### Phase 1: Small startup teams (3-8 devs, 1-2 PMs)
- Pain: "My AI agent doesn't know what my teammate decided yesterday"
- Value: Immediate agent improvement from shared decisions

### Phase 2: Mid-size engineering orgs (20-50 devs, multiple squads)
- Pain: Onboarding, cross-squad context loss, PMs can't track why things were built a certain way
- Value: Institutional knowledge that survives team changes

## AI Agent Support

### Phase 1: Claude Code only
- Deepest integration via hooks (SessionStart, Stop)
- Richest session data for decision extraction

### Future: Agent-agnostic
- MCP server can be added later if mid-session search proves necessary
- Other agents can integrate via the CLI directly

## Tech Stack

| Component | Technology | Rationale |
|---|---|---|
| CLI | Go | Fast startup (~5ms), low memory (~10-15MB), single binary distribution |
| Local index | SQLite + FTS5 | Zero infrastructure, sub-ms queries, disposable cache. sqlite-vec deferred to post-MVP |
| Decision extraction | Claude API (Haiku for cost) | Best at understanding session context and extracting structured decisions |
| macOS app (post-MVP) | Swift/SwiftUI | Native performance, natural fit for macOS |
| Decisions repo | Git + Markdown | Version controlled, inspectable, zero infrastructure |
| Linear | REST API (pull-based) | On-demand fetching, no webhook infrastructure |

## Installation & Onboarding

```bash
brew install backstory

backstory init
# → Creates a decisions repo from template (or connects to existing one)
# → Configures Claude Code hooks in settings.json
# → Done — capture starts on next session
```

One-time setup, ~2 minutes. The macOS app is a separate download for PMs.

## MVP Scope

**In scope:**
- Go CLI with all commands (init, sync, search, index, inject, capture, status, edit)
- Claude Code hooks (SessionStart + Stop)
- Decisions repo template with config schema
- Local SQLite index with FTS5
- Linear integration (pull-based)
- Decision extraction via Claude API
- Pending queue for offline/failure resilience

**Deferred to post-MVP:**
- macOS app for PMs (PMs can use the CLI or edit markdown directly for now)
- Semantic search via embeddings (FTS5 keyword search is sufficient for v1)
- Slack integration (Linear issues carry most PM context)
- Agent-agnostic support / MCP server
- Auto-archive staleness (manual `backstory status` flags stale decisions)

## Open Questions

1. **Decision extraction prompt engineering** — the core AI challenge. How to reliably distinguish real decisions from debugging noise. Needs iteration with real session transcripts. Consider: what does Claude Code's Stop hook actually provide as session data?
2. **macOS app architecture** — menubar app with chat window vs. full app with browse/search/timeline. Depends on PM feedback after CLI MVP ships.
3. **Pricing model** — open source CLI + paid macOS app? Fully open source with paid team features? Cloud-hosted option for teams that don't want to self-manage?
4. **Embedding strategy** — when semantic search is needed post-MVP: local model (increases binary size to ~80MB+) vs. Claude API embeddings (requires network for indexing). FTS5 buys time to decide.
