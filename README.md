# Backstory

Shared team memory for AI coding agents.

Backstory automatically captures the reasoning behind code decisions during AI coding sessions and makes that context available to every developer's agent on the team.

## Install

Build from source:

    go install github.com/backstory-team/backstory/cmd/backstory@latest

Or clone and build:

    git clone https://github.com/backstory-team/backstory.git
    cd backstory
    make build

## Quick Start

    # Create a new decisions repo
    backstory init --path my-team-decisions

    # Or join an existing team's repo
    backstory init --connect git@github.com:my-team/decisions.git

    # Set the environment variable
    export BACKSTORY_REPO=/path/to/my-team-decisions

    # Add API keys
    cat > /path/to/my-team-decisions/.backstory/config.local.yml << EOF
    claude_api_key: sk-ant-...
    linear_api_key: lin_api_...
    EOF

    # Build the search index
    backstory index

    # Search decisions
    backstory search "payment processing"

    # Check status
    backstory status

## Claude Code Integration

Add hooks to your Claude Code settings (`.claude/settings.json`):

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
            "command": "backstory capture --transcript $TRANSCRIPT_PATH"
          }
        ]
      }
    }

## How It Works

1. **Session starts** — Backstory pulls the latest decisions and injects relevant context
2. **You code** — Your agent makes decisions with full team context
3. **Session ends** — Backstory extracts decisions and asks you to confirm sharing
4. **Teammate starts a session** — Their agent already knows what you decided

## Commands

| Command | Description |
|---------|-------------|
| `backstory init` | Create or connect to a decisions repo |
| `backstory sync` | Pull latest, process pending, rebuild index |
| `backstory index` | Rebuild the local search index |
| `backstory search <query>` | Search decisions by keyword |
| `backstory inject` | Output relevant context as XML (used by hooks) |
| `backstory capture` | Extract decisions from session (used by hooks) |
| `backstory status` | Show repo status and pending items |
| `backstory edit <file>` | Edit a decision file |

## License

MIT
