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
│   ├── config.yml        ← linked code repos, team settings
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
anchor_granularity: directory | file | function
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
2. `backstory inject` — reads current code repo path, queries SQLite index for decisions anchored to relevant paths, outputs them as context injected into the session

**Stop hook:**
1. Receives session summary
2. Calls Claude API to extract candidate decisions (the hard AI problem)
3. Presents quick confirmation prompt to the developer:

```
Backstory captured from this session:
  1. [x] Chose SQS over direct invocation for vendor API (rate limit 100 req/s)
  2. [x] Added exponential backoff for Stripe webhooks (can delay up to 5min)
  3. [ ] Fixed flaky test timing issue

[Share 1,2 to team] [Edit] [Dismiss all]
```

4. Confirmed items are written as markdown files and committed/pushed to the decisions repo

**Default is "share"** — items come pre-checked. Developer only acts to remove.

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
- When anchored code changes substantially, the decision is flagged as potentially stale
- Decisions not retrieved by any agent in N months are auto-archived
- Decisions retrieved and then contradicted by agent actions are flagged for review

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
| CLI | Go | Fast startup (~5ms), low memory (~10MB), single binary distribution |
| Local index | SQLite + FTS5 + sqlite-vec | Zero infrastructure, sub-ms queries, disposable cache |
| Decision extraction | Claude API | Best at understanding session context and extracting structured decisions |
| macOS app | Swift/SwiftUI | Native performance, natural fit for macOS |
| Decisions repo | Git + Markdown | Version controlled, inspectable, zero infrastructure |
| Linear/Slack | REST APIs (pull-based) | On-demand fetching, no webhook infrastructure |

## Installation & Onboarding

```bash
brew install backstory

backstory init
# → Creates a decisions repo from template (or connects to existing one)
# → Configures Claude Code hooks in settings.json
# → Done — capture starts on next session
```

One-time setup, ~2 minutes. The macOS app is a separate download for PMs.

## Open Questions

1. **Embedding model for semantic search** — local model (e.g., all-MiniLM) vs. Claude API embeddings? Tradeoff: local is free but less accurate; API is better but adds latency and cost to indexing.
2. **Conflict resolution** — two devs push decisions at the same time. Git handles this for different files, but what about the index? Rebuild on every pull?
3. **Decision extraction quality** — the core AI challenge. How to reliably distinguish real decisions from debugging noise, temporary workarounds, and irrelevant session activity. Needs prompt engineering and possibly fine-tuning.
4. **macOS app scope** — MVP could be a simple menubar app with a chat window, or a full-featured app with browse/search/timeline views.
5. **Pricing model** — open source CLI + paid macOS app? Fully open source with paid team features? Cloud-hosted option for teams that don't want to self-manage?
