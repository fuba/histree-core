# sample.zshrc
# これは zsh-histree の設定例です。
# 以下の行をあなたの .zshrc に追加するか、参考にしてください。

# zsh-histree のパス（リポジトリの場所に合わせて変更してください）
export ZSH_histree_DIR="$HOME/path/to/zsh-histree"

# ログファイルの保存先（必要に応じて変更可能）
export ZSH_PWD_LOG_DIR="${HOME}/.zsh_pwd_log"

# zsh-histree の読み込み
source "$ZSH_histree_DIR/histree.zsh"

# これ以降、通常の zsh の設定を記述してください。
