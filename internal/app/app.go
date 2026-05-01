package app

import (
	"sync"

	"log-viewer/internal/fileio"
)

// App holds the application-level state.
type App struct {
	mu       sync.Mutex
	readers  []*fileio.FileReader
	TempMgr  *fileio.TempFileManager
}

// New creates a new application state.
func New() *App {
	return &App{
		TempMgr: fileio.NewTempFileManager(),
	}
}

// TrackReader registers a FileReader for cleanup on exit.
func (a *App) TrackReader(r *fileio.FileReader) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.readers = append(a.readers, r)
}

// Cleanup closes all open readers and removes temp files.
func (a *App) Cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, r := range a.readers {
		r.Close()
	}
	a.readers = nil
	a.TempMgr.Cleanup()
}
