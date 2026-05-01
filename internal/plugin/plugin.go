package plugin

import "fyne.io/fyne/v2"

// LinePlugin defines the interface for plugins that can process and render log lines.
// Plugins self-register via init() by calling Register().
type LinePlugin interface {
	// Name returns the display name shown in context menus.
	Name() string

	// Description returns a short description of what the plugin does.
	Description() string

	// CanHandle returns true if the plugin can process the given line.
	CanHandle(line string) bool

	// Render returns a Fyne canvas object displaying the processed line.
	// The returned widget is shown in a popup window or split panel.
	Render(line string) fyne.CanvasObject
}
