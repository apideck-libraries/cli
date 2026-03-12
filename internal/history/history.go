package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/apideck-io/cli/internal/spec"
)

const maxEntries = 100

// Log appends a history entry with atomic write and FIFO rotation.
func Log(entry spec.HistoryEntry) error {
	path := DefaultPath()
	entries := load(path)
	entries = append(entries, entry)
	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}
	data, _ := json.MarshalIndent(entries, "", "  ")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load returns all history entries.
func Load() []spec.HistoryEntry {
	return load(DefaultPath())
}

func load(path string) []spec.HistoryEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []spec.HistoryEntry
	json.Unmarshal(data, &entries)
	return entries
}

// DefaultPath returns ~/.apideck-cli/history.json
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".apideck-cli")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "history.json")
}

// NewEntry creates a history entry from a completed request.
func NewEntry(method, path string, status int, duration time.Duration, serviceID string) spec.HistoryEntry {
	return spec.HistoryEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Method:     method,
		Path:       path,
		Status:     status,
		DurationMs: duration.Milliseconds(),
		ServiceID:  serviceID,
	}
}
