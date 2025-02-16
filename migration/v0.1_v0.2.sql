BEGIN TRANSACTION;

-- Create new table with the updated schema
CREATE TABLE history_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    directory TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    exit_code INTEGER NOT NULL,
    hostname TEXT NOT NULL,
    process_id INTEGER NOT NULL
);

-- Copy data from the old table to the new table, extracting hostname and process_id from session_label
INSERT INTO history_new (id, command, directory, timestamp, exit_code, hostname, process_id)
SELECT 
    id,
    command,
    directory,
    timestamp,
    exit_code,
    COALESCE(
        CASE 
            WHEN instr(session_label, '@') > 0 
            THEN substr(session_label, 1, instr(session_label, '@') - 1)
        END,
        'unknown'
    ) as hostname,
    COALESCE(
        CASE 
            WHEN instr(session_label, '@') > 0 
            THEN CAST(
                substr(
                    session_label,
                    instr(session_label, '@') + 1,
                    length(session_label)
                ) AS INTEGER
            )
        END,
        0
    ) as process_id
FROM history;

-- Drop the old table
DROP TABLE history;

-- Rename the new table to the original name
ALTER TABLE history_new RENAME TO history;

-- Recreate the indexes
CREATE INDEX IF NOT EXISTS idx_history_directory ON history(directory);
CREATE INDEX IF NOT EXISTS idx_history_timestamp_directory ON history(timestamp, directory);

COMMIT;