package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type HistoryEntry struct {
	Command   string    `json:"command"`
	Directory string    `json:"directory"`
	Timestamp time.Time `json:"timestamp"`
}

type OutputFormat string

const (
	FormatJSON     OutputFormat = "json"
	FormatReadable OutputFormat = "readable"
)

func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// パフォーマンス最適化のためのPRAGMAを設定
	_, err = db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA temp_store = MEMORY;
		PRAGMA cache_size = -2000;
	`)
	if err != nil {
		return nil, err
	}

	// トランザクション内でテーブルとインデックスを作成
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// テーブル作成
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			directory TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	// インデックスの作成
	// directory検索用のインデックス（LIKE検索の前方一致に効果的）
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_history_directory ON history(directory)`)
	if err != nil {
		return nil, err
	}

	// timestamp + directory の複合インデックス（ORDER BYとWHERE句の両方に効果的）
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_history_timestamp_directory ON history(timestamp, directory)`)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return db, nil
}

func addEntry(db *sql.DB, entry *HistoryEntry) error {
	_, err := db.Exec(
		"INSERT INTO history (command, directory, timestamp) VALUES (?, ?, ?)",
		entry.Command,
		entry.Directory,
		entry.Timestamp,
	)
	return err
}

func getEntries(db *sql.DB, limit int, currentDir string) ([]HistoryEntry, error) {
	// 初期配列サイズを指定して、再アロケーションを減らす
	entries := make([]HistoryEntry, 0, limit)

	query := `
		SELECT command, directory, timestamp 
		FROM history 
		WHERE directory LIKE ? || '%'
		ORDER BY timestamp ASC, directory
		LIMIT ?
	`

	// より大きなバッファサイズでクエリを実行
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// SQLiteのページサイズを設定
	_, err = tx.Exec("PRAGMA page_size = 4096")
	if err != nil {
		return nil, err
	}

	// 結果セットのバッファサイズを設定
	rows, err := tx.Query(query, currentDir, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry HistoryEntry
		err := rows.Scan(&entry.Command, &entry.Directory, &entry.Timestamp)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entries, nil
}

func writeEntries(entries []HistoryEntry, w io.Writer, format OutputFormat) error {
	bufW := bufio.NewWriterSize(w, 8192)

	for _, entry := range entries {
		switch format {
		case FormatJSON:
			if err := json.NewEncoder(bufW).Encode(entry); err != nil {
				return err
			}
		case FormatReadable:
			// JSONっぽい文字列をエスケープ
			command := entry.Command
			if strings.HasPrefix(command, "{") && strings.HasSuffix(command, "}") {
				command = fmt.Sprintf("%q", command)
			}

			if _, err := fmt.Fprintf(bufW, "%s [%s] %s\n",
				entry.Timestamp.Format(time.RFC3339),
				entry.Directory,
				command); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown output format: %s", format)
		}
	}

	return bufW.Flush()
}

func main() {
	dbPath := flag.String("db", "", "Path to SQLite database (required)")
	action := flag.String("action", "", "Action to perform: add or get")
	limit := flag.Int("limit", 100, "Number of entries to retrieve")
	currentDir := flag.String("dir", "", "Current directory for filtering entries")
	format := flag.String("format", string(FormatReadable), "Output format: json or readable")
	flag.Parse()

	if *dbPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -db parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	db, err := initDB(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch *action {
	case "add":
		// 標準入力から全体を読み取る
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, os.Stdin); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read command from stdin: %v\n", err)
			os.Exit(1)
		}
		cmd := buf.String()
		// 最後の改行を削除
		cmd = strings.TrimRight(cmd, "\n")

		dir := *currentDir
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get current directory: %v\n", err)
				os.Exit(1)
			}
		}

		entry := HistoryEntry{
			Command:   cmd,
			Directory: dir,
			Timestamp: time.Now().UTC(),
		}

		if err := addEntry(db, &entry); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add entry: %v\n", err)
			os.Exit(1)
		}

	case "get":
		entries, err := getEntries(db, *limit, *currentDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get entries: %v\n", err)
			os.Exit(1)
		}

		if err := writeEntries(entries, os.Stdout, OutputFormat(*format)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write entries: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", *action)
		os.Exit(1)
	}
}
