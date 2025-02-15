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
typeset -g _HISTREE_LAST_CMD
typeset -g _HISTREE_LAST_EXIT_CODE
_HISTREE_START_TIME=$(date +"%Y%m%d-%H%M%S")
_HISTREE_SESSION_LABEL="${HOST:=$(hostname)}:${_HISTREE_START_TIME}:$$"

# Function to add a command to history
_histree_add_command() {
    local cmd="$_HISTREE_LAST_CMD"
    local exit_code="$_HISTREE_LAST_EXIT_CODE"

    # If the command is empty, do not record it
    if [[ -z "$cmd" ]]; then
        return
    fi

    echo "$cmd" | $HISTREE_BIN -db "$HISTREE_DB" -action add -dir "$PWD" \
        -session "$_HISTREE_SESSION_LABEL" \
        -exit "$exit_code"
}

# Function to capture the last command
_histree_preexec() {
    _HISTREE_LAST_CMD="$1"
}

# Function to capture the last exit code
_histree_precmd() {
    _HISTREE_LAST_EXIT_CODE="$?"
    _histree_add_command
}

# Hook into zsh pre-execution and pre-command
autoload -Uz add-zsh-hook
add-zsh-hook preexec _histree_preexec
add-zsh-hook precmd _histree_precmd

# Function to display history
histree() {
    local format="simple"
    local verbose=false
    local json=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                format="verbose"
                verbose=true
                shift
                ;;
            -json|--json)
                format="json"
                json=true
                shift
                ;;
            *)
                shift
                ;;
        esac
    done

    $HISTREE_BIN -db "$HISTREE_DB" -action get -limit "$HISTREE_LIMIT" -dir "$PWD" -format "$format"
}

# Add aliases for showing history
alias histree='histree'
