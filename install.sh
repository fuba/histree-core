#!/usr/bin/env bash
# install.sh
# zsh-dirstory の自動インストールスクリプト
# ・dirstory.zsh を指定ディレクトリにコピー
# ・.zshrc に source 設定を追加（既に設定済みの場合は追加しません）

set -e

# インストール先ディレクトリ（このスクリプトがあるディレクトリを基準とする）
INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="${HOME}/.zsh-dirstory"

echo "Installing zsh-dirstory to ${TARGET_DIR} ..."
mkdir -p "${TARGET_DIR}"
cp "${INSTALL_DIR}/dirstory.zsh" "${TARGET_DIR}/"

# .zshrc に設定追加
ZSHRC="${HOME}/.zshrc"
SOURCE_LINE="source ${TARGET_DIR}/dirstory.zsh"

if grep -qF "$SOURCE_LINE" "${ZSHRC}"; then
  echo "Your .zshrc already sources zsh-dirstory."
else
  echo "" >> "${ZSHRC}"
  echo "# Load zsh-dirstory" >> "${ZSHRC}"
  echo "$SOURCE_LINE" >> "${ZSHRC}"
  echo "Added source line to ${ZSHRC}."
fi

echo "Installation complete. Please restart your terminal or run 'source ~/.zshrc' to activate zsh-dirstory."
