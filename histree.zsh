#!/usr/bin/env zsh

HISTREE_HOME="${0:A:h}"
HISTREE_BIN="$HISTREE_HOME/bin/histree"
HISTREE_DB="${HISTREE_DB:-$HOME/.histree.db}"
HISTREE_LIMIT="${HISTREE_LIMIT:-100}"

# Ensure the binary exists
if [[ ! -x $HISTREE_BIN ]]; then
    echo "Building histree..."
    (cd "$HISTREE_HOME" && go build -o bin/histree ./cmd/histree)
fi

_histree_add_command() {
    local cmd="$1"
    echo "$cmd" | $HISTREE_BIN -db "$HISTREE_DB" -action add -dir "$PWD"
}

_histree_show_history() {
    $HISTREE_BIN -db "$HISTREE_DB" -action get -limit "$HISTREE_LIMIT" -dir "$PWD" -format readable
}

_histree_show_history_json() {
    $HISTREE_BIN -db "$HISTREE_DB" -action get -limit "$HISTREE_LIMIT" -dir "$PWD" -format json
}

# Hook into zsh
autoload -Uz add-zsh-hook
add-zsh-hook preexec _histree_add_command

# Add aliases for showing history
alias histree=_histree_show_history
alias histree-json=_histree_show_history_json
