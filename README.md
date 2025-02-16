# histree-core

A command-line tool that provides the core functionality for storing and retrieving shell command history with directory context in SQLite database.

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

1. Clone this repository:
    ```sh
    git clone https://github.com/fuba/histree-core.git
    ```

2. Run the following command to build and install histree-core:
    ```sh
    make install
    ```

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

## Requirements

- Go (for building the binary)
- SQLite

## License

This project is licensed under the MIT License.
