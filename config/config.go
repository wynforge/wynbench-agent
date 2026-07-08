// Package config persists Wynbench's connection configuration to a local
// JSON file so it survives agent restarts, and supports export/import of
// the full configuration bundle.
//
// The config file lives at:
//
//	Windows: %USERPROFILE%\wynbench\config\connections.json
//	Linux/macOS: $HOME/wynbench/config/connections.json
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/wynforge/wynbench-agent/core"
)

const fileName = "connections.json"

// Snapshot is the full exportable/importable configuration bundle.
type Snapshot struct {
	Connections []core.Connection `json:"connections"`
	Workflows   []core.Workflow   `json:"workflows,omitempty"`
}

// Dir returns the OS-specific directory where Wynbench persists its config:
// %USERPROFILE%\wynbench\config on Windows, $HOME/wynbench/config elsewhere.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "wynbench", "config"), nil
}

// Path returns the full path to the persisted connections file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Load reads the persisted Snapshot from disk. If no config file exists yet,
// it returns a zero-value Snapshot without error.
func Load() (Snapshot, error) {
	path, err := Path()
	if err != nil {
		return Snapshot{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Snapshot{}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

// Save writes the Snapshot to disk as indented JSON, creating the config
// directory if it does not already exist.
func Save(snap Snapshot) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path, err := Path()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
