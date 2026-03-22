# Backstory Plugin for Claude Code

Shared team memory — auto-captures the reasoning behind code decisions and makes it available to every session.

## Setup

1. Set your decisions repo path:
   ```
   export BACKSTORY_REPO=/path/to/your-team-decisions
   ```

That's it. No API key needed — the plugin uses Claude Code itself to extract decisions.

## What it does

- **Session start:** Syncs latest decisions and injects relevant context
- **During session:** Search and add decisions via MCP tools
- **Session end:** Claude Code reviews the conversation and saves any decisions via the `backstory_save` tool

## MCP Tools

| Tool | Description |
|------|-------------|
| `backstory_search` | Search team decisions by keyword |
| `backstory_add` | Add a new team decision |
| `backstory_save` | Save extracted decisions to the team repo |
| `backstory_status` | Show repo status and counts |
