package fileio

import (
	"fmt"
	"os"
	"sync"
)

// TempFileManager tracks temporary files created during filtering,
// and cleans them up on application exit.
type TempFileManager struct {
	mu    sync.Mutex
	files []string
}

// NewTempFileManager creates a new temp file manager.
func NewTempFileManager() *TempFileManager {
	return &TempFileManager{}
}

// CreateTempFile writes lines to a new temporary file in the OS temp directory.
// Returns the path to the created file.
func (m *TempFileManager) CreateTempFile(lines []string) (string, error) {
	f, err := os.CreateTemp("", "logviewer-filter-*.log")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			os.Remove(f.Name())
			return "", fmt.Errorf("write temp file: %w", err)
		}
	}

	path := f.Name()

	m.mu.Lock()
	m.files = append(m.files, path)
	m.mu.Unlock()

	return path, nil
}

// Cleanup removes all tracked temporary files.
func (m *TempFileManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, path := range m.files {
		os.Remove(path)
	}
	m.files = nil
}
