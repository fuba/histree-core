# zsh-dirstory

`zsh-dirstory` は、実行したコマンドの履歴を、実行時のディレクトリ情報とともに記録し、ディレクトリ階層に沿った「物語（story）」として参照できる zsh プラグインです。  
各ディレクトリごとにログファイルを管理し、コマンド実行後の結果（終了ステータス）や、改行を含むコマンドも安全に記録します。

## 特徴

- **ログ記録フォーマット**  
  各レコードは以下の形式（4行）で記録されます。  
  1. `timestamp<TAB>ulid<TAB>dir`  
  2. `result status`  
  3. `command` （改行などを含む場合は base64 エンコードして 1 行にまとめています）  
  4. 空行（レコード区切り）

- **ディレクトリごとの管理**  
  ログは `$HOME/.zsh_pwd_log` 以下に、各ディレクトリごとにファイル（例：`home_user_project.log`）として保存されます。  
  各ファイルは最大 10,000 行に自動的に制限されます。

- **簡単な抽出コマンド**  
  `dirstory` コマンドを実行するだけで、現在のディレクトリ以下で実行されたコマンド履歴を表示できます。

## インストール

1. 本リポジトリをクローンまたはダウンロードしてください。
```sh
  git clone https://github.com/your-username/zsh-dirstory.git
  ```
2. あなたの .zshrc に dirstory.zsh を読み込む設定を追加します。例:
```sh
# .zshrc の末尾に追加
source /path/to/zsh-dirstory/dirstory.zsh
```
※ リポジトリを ~/zsh-dirstory に配置した場合:
```sh
echo "source ~/zsh-dirstory/dirstory.zsh" >> ~/.zshrc
```
もしくは、同梱の install.sh を実行すると自動で設定が追加されます。

## 使い方
- シェルを再起動するか、.zshrc を再読み込みしてください。
- コマンド実行後、ログが自動で記録されます。
- 現在のディレクトリ以下で実行されたコマンド履歴を確認するには、ターミナルで以下を実行してください。
  ```sh
  dirstory
  ```

## カスタマイズ
- ログファイルの保存先は、環境変数 ZSH_PWD_LOG_DIR で変更可能です。デフォルトは $HOME/.zsh_pwd_log です。
- ULID の生成には、システムに ulid コマンドがある場合はそれを使用し、なければ簡易版を使用します。
