# histree-zsh

**histree-zsh** is a zsh plugin that logs your command history along with the execution directory context, allowing you to explore a hierarchical narrative of your shell activity.

## Features

- **Log Record Format**  
  Each log record is stored in a 4-line format:
  1. `timestamp<TAB>ulid<TAB>dir`
  2. `result status`
  3. `command` (if the command contains newlines, it is base64-encoded to keep it on one line)
  4. An empty line (record separator)

- **Directory-Based Management**  
  Logs are stored in files under `$HOME/.zsh_pwd_log`, with one file per directory.  
  Each file is automatically capped at a maximum of 10,000 lines.

- **Easy Retrieval**  
  Simply run the `histree` command to display the command history executed in the current directory (including subdirectories).

## Installation

1. Clone or download this repository:
    ```sh
    git clone https://github.com/your-username/histree-zsh.git
    ```

2. Add the following line to your .zshrc to source the plugin:
    ```sh
    source /path/to/histree-zsh/histree.zsh
    ```
    For example, if you place the repository at ~/histree-zsh:
    ```sh
    echo "source ~/histree-zsh/histree.zsh" >> ~/.zshrc
    ```

3. Alternatively, run the included install.sh script to automatically add the source line to your .zshrc.

## Usage
- Restart your terminal or source your .zshrc to activate histree-zsh.
- After each command execution, a log entry is automatically recorded.
- To display the command history for the current directory and its subdirectories, run:

    ```sh
    histree
    ```
## Customization
- You can change the log storage directory by setting the ZSH_PWD_LOG_DIR environment variable. The default is $HOME/.zsh_pwd_log.
- The ULID is generated using the system’s ulid command if available, or a fallback simple version otherwise.


## License
This project is licensed under the MIT License. 
