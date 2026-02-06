#!/bin/bash
# Install histree hook for Claude Code
# This script sets up a PostToolUse hook to record Bash commands to histree

set -e

CLAUDE_DIR="$HOME/.claude"
HOOKS_DIR="$CLAUDE_DIR/hooks"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"
HOOK_SCRIPT="$HOOKS_DIR/post-bash-histree.sh"

# Detect histree-core path
if [ -x "$HOME/.histree-zsh/bin/histree-core" ]; then
    HISTREE_CORE="$HOME/.histree-zsh/bin/histree-core"
elif command -v histree-core &> /dev/null; then
    HISTREE_CORE=$(command -v histree-core)
else
    echo "Error: histree-core not found"
    echo "Please install histree-core first"
    exit 1
fi

# Database path
HISTREE_DB="$HOME/.histree.db"

echo "Installing Claude Code hook for histree..."
echo "  histree-core: $HISTREE_CORE"
echo "  database: $HISTREE_DB"

# Create hooks directory
mkdir -p "$HOOKS_DIR"

# Create hook script (use heredoc with variable expansion for paths)
cat > "$HOOK_SCRIPT" << HOOKEOF
#!/bin/bash
# Hook script to record Claude Code's Bash commands to histree
# This script is called by Claude Code after each Bash command execution

# Read JSON input from stdin
INPUT=\$(cat)

# Extract command, exit code, and working directory
COMMAND=\$(echo "\$INPUT" | jq -r '.tool_input.command // empty')
EXIT_CODE=\$(echo "\$INPUT" | jq -r '.tool_response.exit_code // 0')
CWD=\$(echo "\$INPUT" | jq -r '.cwd // empty')
SESSION_ID=\$(echo "\$INPUT" | jq -r '.session_id // "unknown"')

# Skip if command is empty
if [ -z "\$COMMAND" ]; then
    exit 0
fi

# Get hostname
HOSTNAME=\$(hostname)

# Path to histree-core and database
HISTREE_CORE="$HISTREE_CORE"
HISTREE_DB="$HISTREE_DB"

# Record to histree (using session_id hash as pid to group related commands)
# Use a pseudo-PID based on session_id to group Claude Code commands
PSEUDO_PID=\$(echo "\$SESSION_ID" | cksum | awk '{print \$1 % 100000 + 900000}')

echo "\$COMMAND" | "\$HISTREE_CORE" \\
    -db "\$HISTREE_DB" \\
    -action add \\
    -hostname "\${HOSTNAME}:claude" \\
    -pid "\$PSEUDO_PID" \\
    -dir "\$CWD" \\
    -exit "\$EXIT_CODE" \\
    2>/dev/null

# Always exit 0 to not block Claude Code
exit 0
HOOKEOF

# Make executable
chmod +x "$HOOK_SCRIPT"
echo "  Created hook script: $HOOK_SCRIPT"

# Update settings.json
if [ ! -f "$SETTINGS_FILE" ]; then
    # Create new settings file
    cat > "$SETTINGS_FILE" << EOF
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$HOOK_SCRIPT",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
EOF
    echo "  Created settings file: $SETTINGS_FILE"
else
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo "Error: jq is required to update existing settings.json"
        echo "Please install jq or manually add the hook configuration"
        exit 1
    fi

    # Check if hooks already configured
    if jq -e '.hooks.PostToolUse' "$SETTINGS_FILE" > /dev/null 2>&1; then
        # Check if histree hook already exists
        if jq -e '.hooks.PostToolUse[] | select(.hooks[].command | contains("post-bash-histree"))' "$SETTINGS_FILE" > /dev/null 2>&1; then
            echo "  Hook already configured in settings.json"
        else
            # Add to existing PostToolUse array
            TMP_FILE=$(mktemp)
            jq --arg cmd "$HOOK_SCRIPT" '.hooks.PostToolUse += [{
                "matcher": "Bash",
                "hooks": [{
                    "type": "command",
                    "command": $cmd,
                    "timeout": 5
                }]
            }]' "$SETTINGS_FILE" > "$TMP_FILE"
            mv "$TMP_FILE" "$SETTINGS_FILE"
            echo "  Added hook to existing PostToolUse in settings.json"
        fi
    elif jq -e '.hooks' "$SETTINGS_FILE" > /dev/null 2>&1; then
        # hooks exists but no PostToolUse
        TMP_FILE=$(mktemp)
        jq --arg cmd "$HOOK_SCRIPT" '.hooks.PostToolUse = [{
            "matcher": "Bash",
            "hooks": [{
                "type": "command",
                "command": $cmd,
                "timeout": 5
            }]
        }]' "$SETTINGS_FILE" > "$TMP_FILE"
        mv "$TMP_FILE" "$SETTINGS_FILE"
        echo "  Added PostToolUse hook to settings.json"
    else
        # No hooks at all
        TMP_FILE=$(mktemp)
        jq --arg cmd "$HOOK_SCRIPT" '. + {
            "hooks": {
                "PostToolUse": [{
                    "matcher": "Bash",
                    "hooks": [{
                        "type": "command",
                        "command": $cmd,
                        "timeout": 5
                    }]
                }]
            }
        }' "$SETTINGS_FILE" > "$TMP_FILE"
        mv "$TMP_FILE" "$SETTINGS_FILE"
        echo "  Added hooks configuration to settings.json"
    fi
fi

echo ""
echo "Installation complete!"
echo "Restart Claude Code to activate the hook."
echo ""
echo "Commands will be recorded with hostname suffix ':claude'"
echo "View history: $HISTREE_CORE -db $HISTREE_DB -action get -v | grep :claude"
