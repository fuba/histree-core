# ===================================
# zsh-dirstory
#
# このスクリプトは、コマンド実行時の情報を
# ログファイル（$HOME/.zsh_pwd_log 以下、ディレクトリごとに分割）に記録し、
# dirstory 関数で現在のディレクトリ以下の履歴を表示します。
#
# ログの各レコードは、以下の形式で記録されます:
#   1行目: timestamp<TAB>ulid<TAB>dir
#   2行目: result status
#   3行目: command（改行を含む場合は base64 エンコード）
#   4行目: 空行（レコード区切り）
#
# ※ dirstory 関数実行時に、base64 でエンコードされたコマンドをデコードして出力します。
# ===================================

# -----------------------------------
# ログ保存先ディレクトリの設定
# -----------------------------------
export ZSH_PWD_LOG_DIR="${HOME}/.zsh_pwd_log"
mkdir -p "$ZSH_PWD_LOG_DIR"

# -----------------------------------
# sanitize_dir: ディレクトリ名をファイル名用に変換
# 例: /home/user/project/sub → home_user_project_sub.log
# -----------------------------------
sanitize_dir() {
  local dir="${1#/}"   # 先頭の "/" を除去
  dir=${dir//\//_}      # "/" を "_" に置換
  if [[ -z "$dir" ]]; then
    echo "root"
  else
    echo "$dir"
  fi
}

# -----------------------------------
# generate_ulid: ULID（または代替のユニーク ID）を生成
# -----------------------------------
generate_ulid() {
  if command -v ulid > /dev/null 2>&1; then
    ulid
  else
    # 簡易版 ULID: timestamp と RANDOM を組み合わせたもの
    printf "%s%04X" "$(date +%s)" $(( RANDOM ))
  fi
}

# -----------------------------------
# ログ用一時変数（preexec/ precmd で利用）
# -----------------------------------
__last_log_timestamp=""
__last_log_ulid=""
__last_log_dir=""
__last_log_command=""

# -----------------------------------
# preexec: コマンド実行直前に情報を保持する
# -----------------------------------
preexec() {
  __last_log_timestamp=$(date +%s)
  __last_log_ulid=$(generate_ulid)
  __last_log_dir=$PWD
  __last_log_command=$1
}

# -----------------------------------
# precmd: コマンド実行後にログファイルへ出力する
#
# 出力形式:
#   1行目: timestamp<TAB>ulid<TAB>dir
#   2行目: result status
#   3行目: command（改行を含む場合は base64 エンコード済み）
#   4行目: 空行（レコード区切り）
# -----------------------------------
precmd() {
  local status=$?
  if [[ -n "$__last_log_command" ]]; then
    local sanitized_dir
    sanitized_dir=$(sanitize_dir "$__last_log_dir")
    local log_file="${ZSH_PWD_LOG_DIR}/${sanitized_dir}.log"
    # コマンド部分を base64 エンコード（改行含む場合も 1 行にまとめる）
    local encoded_command
    encoded_command=$(printf "%s" "$__last_log_command" | base64 | tr -d '\n')
    {
      printf "%s\t%s\t%s\n" "$__last_log_timestamp" "$__last_log_ulid" "$__last_log_dir"
      printf "%s\n" "$status"
      printf "%s\n" "$encoded_command"
      printf "\n"
    } >> "$log_file"

    # ログファイルが 10,000 行を超えた場合、最新 10,000 行のみ残す
    if [ -f "$log_file" ]; then
      local line_count
      line_count=$(wc -l < "$log_file")
      if [ "$line_count" -gt 10000 ]; then
        tail -n 10000 "$log_file" > "${log_file}.tmp" && mv "${log_file}.tmp" "$log_file"
      fi
    fi

    unset __last_log_timestamp __last_log_ulid __last_log_dir __last_log_command
  fi
}

# -----------------------------------
# dirstory: 現在のディレクトリ以下で実行されたコマンドを抽出する関数
#
# 出力形式（レコード間は空行で区切られる）:
#   1行目: timestamp<TAB>ulid<TAB>dir
#   2行目: result status
#   3行目: command（base64 で記録されたものをデコードして元の改行を復元）
# -----------------------------------
dirstory() {
  local curr_dir="$PWD"
  for logfile in "$ZSH_PWD_LOG_DIR"/*.log; do
    [ -f "$logfile" ] || continue
    # awk でレコード（空行区切り）単位に処理
    awk -v RS="" -v ORS="\n\n" -v curr_dir="$curr_dir" '
      {
        # 1行目をタブで分割し、3 番目のフィールド（記録された dir）を取得
        split($1, header, "\t");
        if (header[3] == curr_dir || index(header[3], curr_dir "/") == 1) {
          print $0;
        }
      }
    ' "$logfile" | while IFS= read -r record; do
      # 各レコードは以下の行になっている前提：
      # 1行目: header (timestamp<TAB>ulid<TAB>dir)
      # 2行目: result status
      # 3行目: encoded command（base64）
      local header status encoded_cmd
      header=$(printf "%s" "$record" | sed -n '1p')
      status=$(printf "%s" "$record" | sed -n '2p')
      encoded_cmd=$(printf "%s" "$record" | sed -n '3p')
      # encoded_cmd を base64 デコードして元のコマンド（改行付き）に戻す
      local decoded_cmd
      decoded_cmd=$(printf "%s" "$encoded_cmd" | base64 -d 2>/dev/null)
      printf "%s\n%s\n%s\n\n" "$header" "$status" "$decoded_cmd"
    done
  done
}

