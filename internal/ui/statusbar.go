package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// StatusBar displays file information at the bottom of the window.
type StatusBar struct {
	widget.BaseWidget

	lineCount    *widget.Label
	fileSize     *widget.Label
	selectedLine *widget.Label
	container    fyne.CanvasObject
}

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	s := &StatusBar{
		lineCount:    widget.NewLabel("Lines: 0"),
		fileSize:     widget.NewLabel("Size: 0 B"),
		selectedLine: widget.NewLabel("Ln: -"),
	}
	return s
}

// Container returns the status bar as a renderable container.
func (s *StatusBar) Container() fyne.CanvasObject {
	return newHBox(s.lineCount, s.fileSize, s.selectedLine)
}

// Update refreshes the status bar with new values.
func (s *StatusBar) Update(lines int, size int64, selected int) {
	s.lineCount.SetText(fmt.Sprintf("Lines: %d", lines))
	s.fileSize.SetText(fmt.Sprintf("Size: %s", formatSize(size)))
	if selected >= 0 {
		s.selectedLine.SetText(fmt.Sprintf("Ln: %d", selected+1))
	} else {
		s.selectedLine.SetText("Ln: -")
	}
}

func formatSize(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
