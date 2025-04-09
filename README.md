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

- **Directory Path Updates**  
  When you move or rename directories, you can update all related history entries:
  - Updates both exact path matches and subdirectory paths
  - Preserves your command history context when reorganizing your filesystem
  - Handles relative paths automatically

- **Shell Context Tracking**
  Each command is stored with its execution context:
  - Hostname of the machine
  - Shell process ID
  - Exit code
  - Timestamp (stored in UTC, displayed in local timezone)
  - Working directory

## Installation

### Using as a Command-Line Tool

#### Go Install (recommended)
```sh
go install github.com/fuba/histree-core/cmd/histree-core@latest
```

#### Building from Source
```sh
git clone https://github.com/fuba/histree-core.git
cd histree-core
make build
```

Shell-specific implementations are also available—for example, install [histree-zsh](https://github.com/fuba/histree-zsh) for Zsh.

### Using as a Library

Add histree-core to your Go project:

```sh
go get github.com/fuba/histree-core
```

Then import the package in your Go code:

```go
import "github.com/fuba/histree-core/pkg/histree"
```

## Command Line Options

```sh
-db string      Path to SQLite database (required)
-action string  Action to perform: add, get, or update-path
-dir string     Current directory for filtering entries
-format string  Output format: json, simple, or verbose (default "simple")
-limit int      Number of entries to retrieve (default 100)
-hostname       Hostname for command history (required for add action)
-pid            Process ID of the shell (required for add action)
-exit int       Exit code of the command
-old-path       Old directory path (required for update-path action)
-new-path       New directory path (required for update-path action)
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
2024-02-15T15:04:30 [/path/to/directory] [exit_code] command
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

## Library Usage

The histree-core package can be used as a library in your Go applications:

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fuba/histree-core/pkg/histree"
)

func main() {
	// Open or create a history database
	db, err := histree.OpenDB("path/to/history.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Add a new entry
	entry := &histree.HistoryEntry{
		Command:   "ls -la",
		Directory: "/home/user/projects",
		Timestamp: time.Now().UTC(),
		ExitCode:  0,
		Hostname:  "myhost",
		ProcessID: 1234,
	}
	if err := db.AddEntry(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add entry: %v\n", err)
		os.Exit(1)
	}

	// Retrieve history entries
	entries, err := db.GetEntries(10, "/home/user/projects")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get entries: %v\n", err)
		os.Exit(1)
	}

	// Output entries in verbose format
	if err := histree.WriteEntries(entries, os.Stdout, histree.FormatVerbose); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write entries: %v\n", err)
		os.Exit(1)
	}
	
	// Update directory paths (e.g., after moving directories)
	oldPath := "/home/user/old-project-path"
	newPath := "/home/user/new-project-path"
	count, err := db.UpdatePaths(oldPath, newPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to update paths: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated %d history entries\n", count)
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
2024-02-15T10:30:15 [/home/user/projects/web-app] npm install
2024-02-15T10:31:20 [/home/user/projects/web-app] npm run build
2024-02-15T10:31:45 [/home/user/projects/web-app/dist] ls -la
2024-02-15T10:32:10 [/home/user/projects/web-app] git status

$ histree -json        # View history in JSON format
{"command":"npm install","directory":"/home/user/projects/web-app","timestamp":"2024-02-15T10:30:15Z","hostname":"laptop","process_id":1234}
{"command":"npm run build","directory":"/home/user/projects/web-app","timestamp":"2024-02-15T10:31:20Z","hostname":"laptop","process_id":1234}
{"command":"ls -la","directory":"/home/user/projects/web-app/dist","timestamp":"2024-02-15T10:31:45Z","hostname":"laptop","process_id":1234}
{"command":"git status","directory":"/home/user/projects/web-app","timestamp":"2024-02-15T10:32:10Z","hostname":"laptop","process_id":1234}

$ cd ~/another-project
$ histree -v           # Different directory shows different history
2024-02-15T10:35:00 [/home/user/another-project] vim README.md
2024-02-15T10:35:30 [/home/user/another-project] git add README.md
2024-02-15T10:36:00 [/home/user/another-project] git commit -m "Update README"

$ # Now let's move a directory and update history
$ mv ~/projects/web-app ~/projects/renamed-app
$ histree-core -db ~/.histree.db -action update-path -old-path ~/projects/web-app -new-path ~/projects/renamed-app
Updated 4 entries: /home/user/projects/web-app -> /home/user/projects/renamed-app

$ cd ~/projects/renamed-app
$ histree -v           # History is preserved with the new path
2024-02-15T10:30:15 [/home/user/projects/renamed-app] npm install
2024-02-15T10:31:20 [/home/user/projects/renamed-app] npm run build
2024-02-15T10:31:45 [/home/user/projects/renamed-app/dist] ls -la
2024-02-15T10:32:10 [/home/user/projects/renamed-app] git status
```

This example demonstrates how histree helps track your development workflow across different directories and projects, maintaining the context of your work.

## Requirements

- Go 1.18 or later (for building the binary and using as a library)
- SQLite

## License

This project is licensed under the MIT License.
