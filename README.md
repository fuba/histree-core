# histree-core

A command-line tool that provides the core functionality for storing and retrieving shell command history with directory context in SQLite database.

This project was developed with the assistance of ChatGPT and GitHub Copilot.

## Features

- **SQLite-Based Storage**  
  Command history is stored in a SQLite database, providing reliable and efficient storage with:
  - Optimized indexes for fast directory-based queries
  - Transaction support for data integrity
  - WAL mode for better concurrent access

- **Directory-Aware History**  
  Commands are stored with their execution directory context, allowing you to view history specific to directories.

- **Shell Context Tracking**
  Each command is stored with its execution context:
  - Hostname of the machine
  - Shell process ID
  - Exit code
  - Timestamp (UTC)
  - Working directory

## Installation

Please install using the shell-specific implementation—for example, install [histree-zsh](https://github.com/fuba/histree-zsh) for Zsh.

## Command Line Options

```sh
-db string      Path to SQLite database (required)
-action string  Action to perform: add or get
-dir string     Current directory for filtering entries
-format string  Output format: json, simple, or verbose (default "simple")
-limit int      Number of entries to retrieve (default 100)
-hostname       Hostname for command history (required for add action)
-pid           Process ID of the shell (required for add action)
-exit int       Exit code of the command
-v              Show verbose output (same as -format verbose)
```

## Output Formats

The tool supports three output formats:

1. Simple format (default):
```sh
command
```

2. Verbose format:
```sh
2024-02-15T15:04:30Z [/path/to/directory] [exit_code] command
```

3. JSON format:
```json
{
  "command": "command string",
  "directory": "/path/to/directory",
  "timestamp": "2024-02-15T15:04:30Z",
  "exit_code": 0,
  "hostname": "host",
  "process_id": 1234
}
```

### Example Session of [histree-zsh](https://github.com/fuba/histree-zsh) 

```sh
$ cd ~/projects/web-app
$ npm install           # Installing dependencies
$ npm run build        # Building the project
$ cd dist
$ ls -la              # Checking build output
$ cd ..
$ git status

$ cd ~/another-project
$ vim README.md        # Editing README
$ git add README.md
$ git commit -m "Update README"

$ cd ~/projects/web-app
$ histree -v           # View detailed history in current directory
2024-02-15T10:30:15Z [/home/user/projects/web-app] npm install
2024-02-15T10:31:20Z [/home/user/projects/web-app] npm run build
2024-02-15T10:31:45Z [/home/user/projects/web-app/dist] ls -la
2024-02-15T10:32:10Z [/home/user/projects/web-app] git status

$ histree -json        # View history in JSON format
{"command":"npm install","directory":"/home/user/projects/web-app","timestamp":"2024-02-15T10:30:15Z","session_label":"laptop:20240215-103012:1234"}
{"command":"npm run build","directory":"/home/user/projects/web-app","timestamp":"2024-02-15T10:31:20Z","session_label":"laptop:20240215-103012:1234"}
{"command":"ls -la","directory":"/home/user/projects/web-app/dist","timestamp":"2024-02-15T10:31:45Z","session_label":"laptop:20240215-103012:1234"}
{"command":"git status","directory":"/home/user/projects/web-app","timestamp":"2024-02-15T10:32:10Z","session_label":"laptop:20240215-103012:1234"}

$ cd ~/another-project
$ histree -v           # Different directory shows different history
2024-02-15T10:35:00Z [/home/user/another-project] vim README.md
2024-02-15T10:35:30Z [/home/user/another-project] git add README.md
2024-02-15T10:36:00Z [/home/user/another-project] git commit -m "Update README"
```

This example demonstrates how histree helps track your development workflow across different directories and projects, maintaining the context of your work.

## Requirements

- Go (for building the binary)
- SQLite

## License

This project is licensed under the MIT License.
