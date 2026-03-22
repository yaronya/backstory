#!/bin/bash
if [ -z "$BACKSTORY_REPO" ]; then exit 0; fi
BIN="${CLAUDE_PLUGIN_ROOT}/bin/backstory"
if [ ! -f "$BIN" ]; then exit 0; fi
"$BIN" sync --yes 2>/dev/null
"$BIN" inject 2>/dev/null
