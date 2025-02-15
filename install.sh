#!/usr/bin/env bash
# install.sh
# histree-zsh installation script.
# This script copies histree.zsh to a target directory and appends the source command
# to your .zshrc if it is not already present.

set -e

# Determine the installation directory (based on the script location)
INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="${HOME}/.histree-zsh"

echo "Installing histree-zsh to ${TARGET_DIR} ..."
mkdir -p "${TARGET_DIR}"
cp "${INSTALL_DIR}/histree.zsh" "${TARGET_DIR}/"

# Append the source command to .zshrc if not already present
ZSHRC="${HOME}/.zshrc"
SOURCE_LINE="source ${TARGET_DIR}/histree.zsh"

if grep -qF "$SOURCE_LINE" "${ZSHRC}"; then
  echo "Your .zshrc already sources histree-zsh."
else
  echo "" >> "${ZSHRC}"
  echo "# Load histree-zsh" >> "${ZSHRC}"
  echo "$SOURCE_LINE" >> "${ZSHRC}"
  echo "Added source line to ${ZSHRC}."
fi

echo "Installation complete. Please restart your terminal or run 'source ~/.zshrc' to activate histree-zsh."
