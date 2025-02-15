#!/usr/bin/env zsh

# Set default configuration paths and values
HISTREE_HOME="${0:A:h}"
HISTREE_BIN="$HISTREE_HOME/bin/histree"
HISTREE_DB="${HISTREE_DB:-$HOME/.histree.db}"
HISTREE_LIMIT="${HISTREE_LIMIT:-100}"

# Build the binary if it doesn't exist
if [[ ! -x $HISTREE_BIN ]]; then
    echo "Building histree..."
    (cd "$HISTREE_HOME" && go build -o bin/histree ./cmd/histree)
fi

# Generate unique session label when the plugin is loaded
typeset -g _HISTREE_START_TIME
typeset -g _HISTREE_SESSION_LABEL
_HISTREE_START_TIME=$(date +"%Y%m%d-%H%M%S")
_HISTREE_SESSION_LABEL="${HOST:=$(hostname)}:${_HISTREE_START_TIME}:$$"

# Function to add a command to history
_histree_add_command() {
    local cmd="$1"
    echo "$cmd" | $HISTREE_BIN -db "$HISTREE_DB" -action add -dir "$PWD" -session "$_HISTREE_SESSION_LABEL"
}

# Function to display formatted history
_histree_show_history() {
    $HISTREE_BIN -db "$HISTREE_DB" -action get -limit "$HISTREE_LIMIT" -dir "$PWD" -format readable
}

# Function to display history in JSON format
_histree_show_history_json() {
    $HISTREE_BIN -db "$HISTREE_DB" -action get -limit "$HISTREE_LIMIT" -dir "$PWD" -format json
}

# Hook into zsh pre-execution
autoload -Uz add-zsh-hook
add-zsh-hook preexec _histree_add_command

# Add aliases for showing history
alias histree=_histree_show_history
alias histree-json=_histree_show_history_json
