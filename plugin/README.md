# Backstory Plugin for Claude Code

Shared team memory — auto-captures the reasoning behind code decisions and makes it available to every session.

## Setup

1. Set your decisions repo path:
   ```
   export BACKSTORY_REPO=/path/to/your-team-decisions
   ```

2. Add your API key:
   ```
   echo "claude_api_key: sk-ant-..." >> $BACKSTORY_REPO/.backstory/config.local.yml
   ```

## What it does

- **Session start:** Syncs latest decisions and injects relevant context
- **During session:** Search and add decisions via MCP tools
- **Session end:** Auto-captures decisions from your transcript

## MCP Tools

| Tool | Description |
|------|-------------|
| `backstory_search` | Search team decisions by keyword |
| `backstory_add` | Add a new team decision |
| `backstory_status` | Show repo status and counts |
