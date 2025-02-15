#!/usr/bin/env bash
# install.sh
# histree-zsh installation script.
set -e

# Determine the installation directory (based on the script location)
INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="${HOME}/.histree-zsh"

echo "Installing histree-zsh to ${TARGET_DIR} ..."

# Create target directory and copy files
mkdir -p "${TARGET_DIR}/bin"
cp -r "${INSTALL_DIR}/histree.zsh" "${TARGET_DIR}/"

# Build the Go binary
echo "Building histree binary..."
(cd "${INSTALL_DIR}" && go build -o "${TARGET_DIR}/bin/histree" ./cmd/histree)

# Add configuration to .zshrc if not already present
ZSHRC="${HOME}/.zshrc"
SOURCE_LINE="source ${TARGET_DIR}/histree.zsh"

# Default configurations
DB_CONFIG="export HISTREE_DB=\"\${HOME}/.histree.db\""
LIMIT_CONFIG="export HISTREE_LIMIT=100"

if grep -qF "$SOURCE_LINE" "${ZSHRC}"; then
    echo "Your .zshrc already sources histree-zsh."
else
    echo "" >> "${ZSHRC}"
    echo "# histree-zsh configuration" >> "${ZSHRC}"
    echo "$DB_CONFIG" >> "${ZSHRC}"
    echo "$LIMIT_CONFIG" >> "${ZSHRC}"
    echo "$SOURCE_LINE" >> "${ZSHRC}"
    echo "Added configuration to ${ZSHRC}."
fi

echo "Installation complete. Please restart your terminal or run 'source ~/.zshrc' to activate histree-zsh."
