#!/usr/bin/env bash
# install.sh
# zsh-histree の自動インストールスクリプト
# ・histree.zsh を指定ディレクトリにコピー
# ・.zshrc に source 設定を追加（既に設定済みの場合は追加しません）

set -e

# インストール先ディレクトリ（このスクリプトがあるディレクトリを基準とする）
INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="${HOME}/.zsh-histree"

echo "Installing zsh-histree to ${TARGET_DIR} ..."
mkdir -p "${TARGET_DIR}"
cp "${INSTALL_DIR}/histree.zsh" "${TARGET_DIR}/"

# .zshrc に設定追加
ZSHRC="${HOME}/.zshrc"
SOURCE_LINE="source ${TARGET_DIR}/histree.zsh"

if grep -qF "$SOURCE_LINE" "${ZSHRC}"; then
  echo "Your .zshrc already sources zsh-histree."
else
  echo "" >> "${ZSHRC}"
  echo "# Load zsh-histree" >> "${ZSHRC}"
  echo "$SOURCE_LINE" >> "${ZSHRC}"
  echo "Added source line to ${ZSHRC}."
fi

echo "Installation complete. Please restart your terminal or run 'source ~/.zshrc' to activate zsh-histree."
