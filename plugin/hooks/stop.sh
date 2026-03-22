#!/bin/bash
if [ -z "$BACKSTORY_REPO" ]; then exit 0; fi
BIN="${CLAUDE_PLUGIN_ROOT}/bin/backstory"
if [ ! -f "$BIN" ]; then exit 0; fi
"$BIN" capture --author "${USER:-unknown}" 2>/dev/null
