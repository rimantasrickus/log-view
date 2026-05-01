package main

import (
	"os"

	"fyne.io/fyne/v2/app"

	logapp "log-viewer/internal/app"
	"log-viewer/internal/ui"

	// Register plugins
	_ "log-viewer/internal/plugin/jsonformat"
)

func main() {
	fyneApp := app.NewWithID("com.logviewer.app")
	fyneApp.Settings().SetTheme(ui.NewLogViewerTheme())
	appState := logapp.New()

	mainWindow := ui.NewMainWindow(fyneApp, appState)

	// Open files passed as CLI arguments
	for _, arg := range os.Args[1:] {
		mainWindow.OpenFile(arg)
	}

	// Listen for macOS "Open With" file events
	openCh := initOpenFileHandler()
	go func() {
		for path := range openCh {
			mainWindow.OpenFile(path)
		}
	}()

	// Cleanup on exit
	mainWindow.Window().SetOnClosed(func() {
		appState.Cleanup()
	})

	mainWindow.Show()
}
