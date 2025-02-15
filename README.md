# zsh-histree

**zsh-histree** is a zsh plugin that logs your command history along with the execution directory context, allowing you to explore a hierarchical narrative of your shell activity.

This project was developed with the assistance of ChatGPT and GitHub Copilot.

## Features

- **SQLite-Based Storage**  
  Command history is stored in a SQLite database, providing reliable and efficient storage with:
  - Optimized indexes for fast directory-based queries
  - Transaction support for data integrity
  - WAL mode for better concurrent access

- **Directory-Aware History**  
  Commands are stored with their execution directory context, allowing you to view history specific to your current directory and its subdirectories.

- **Session Tracking**
  Each zsh session is uniquely identified with:
  - Hostname of the machine
  - Session start timestamp
  - Process ID
  These are combined into a human-readable session label (e.g., "hostname:20240215-123456:1234")

- **Smart Command Output**
  - Readable format: `timestamp [directory] (session) command`
  - JSON format for programmatic access
  - Proper handling of multi-line commands
  - Intelligent escaping of special characters

## Installation

1. Clone this repository:
    ```sh
    git clone https://github.com/your-username/zsh-histree.git
    ```

2. Run the installation script:
    ```sh
    ./install.sh
    ```

The install script will:
- Create the necessary directories
- Build the Go binary
- Add the configuration to your .zshrc
- Set up the default database location

## Configuration

### Database Location
By default, the history database is stored at `$HOME/.histree.db`. You can change this location by setting the `HISTREE_DB` environment variable in your `.zshrc`:

```zsh
export HISTREE_DB="$HOME/.config/histree/history.db"
```

### History Limit
The number of history entries to display can be configured using the `HISTREE_LIMIT` environment variable (default: 100):

```zsh
export HISTREE_LIMIT=500
```

### Command Line Options

When using the histree commands directly, the following options are available:

```sh
-db string      Path to SQLite database (required)
-action string  Action to perform: add or get
-dir string     Current directory for filtering entries
-format string  Output format: json or readable (default "readable")
-limit int      Number of entries to retrieve (default 100)
-session string Session label for command history (required for add action)
```

## Usage

After installation, zsh-histree will automatically start logging your commands.

To view command history:
```sh
histree          # View formatted history for current directory and subdirectories
histree-json     # View history in JSON format
```

The history display includes:
- Timestamp of command execution
- Directory where the command was run (`[directory]`)
- Session identifier (`(hostname:YYYYMMDD-HHMMSS:pid)`)
- The command itself

## Requirements

- Zsh
- Go (for building the binary)
- SQLite

## License

This project is licensed under the MIT License.
