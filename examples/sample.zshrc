# sample.zshrc
# これは zsh-dirstory の設定例です。
# 以下の行をあなたの .zshrc に追加するか、参考にしてください。

# zsh-dirstory のパス（リポジトリの場所に合わせて変更してください）
export ZSH_DIRSTORY_DIR="$HOME/path/to/zsh-dirstory"

# ログファイルの保存先（必要に応じて変更可能）
export ZSH_PWD_LOG_DIR="${HOME}/.zsh_pwd_log"

# zsh-dirstory の読み込み
source "$ZSH_DIRSTORY_DIR/dirstory.zsh"

# これ以降、通常の zsh の設定を記述してください。
