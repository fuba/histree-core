package histree

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WriteEntries writes history entries to the provided writer using the specified format
func WriteEntries(entries []HistoryEntry, w io.Writer, format OutputFormat) error {
	bufW := bufio.NewWriterSize(w, 8192)

	for _, entry := range entries {
		switch format {
		case FormatJSON:
			if err := json.NewEncoder(bufW).Encode(entry); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		case FormatSimple:
			if _, err := fmt.Fprintf(bufW, "%s\n", entry.Command); err != nil {
				return fmt.Errorf("failed to write entry: %w", err)
			}
		case FormatVerbose:
			command := entry.Command
			if strings.HasPrefix(command, "{") && strings.HasSuffix(command, "}") {
				command = fmt.Sprintf("%q", command)
			}

			exitStatus := ""
			if entry.ExitCode != 0 {
				exitStatus = fmt.Sprintf(" [%d]", entry.ExitCode)
			}

			// Convert UTC time to local timezone
			localTime := entry.Timestamp.Local()

			if _, err := fmt.Fprintf(bufW, "%s [%s]%s %s\n",
				localTime.Format("2006-01-02T15:04:05"),
				entry.Directory,
				exitStatus,
				command); err != nil {
				return fmt.Errorf("failed to write entry: %w", err)
			}
		default:
			return fmt.Errorf("unknown output format: %s", format)
		}
	}

	if err := bufW.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}
	return nil
}
