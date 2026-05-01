package ui

import (
	"io"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"log-viewer/internal/app"
	"log-viewer/internal/fileio"
)

// MainWindow holds the main application window state.
type MainWindow struct {
	window     fyne.Window
	appState   *app.App
	tabs       *container.DocTabs
	logTabs    []*LogTab
	welcomeTab *container.TabItem
}

// NewMainWindow creates and configures the main application window.
func NewMainWindow(fyneApp fyne.App, appState *app.App) *MainWindow {
	win := fyneApp.NewWindow("Log Viewer")
	win.Resize(fyne.NewSize(1400, 800))

	mw := &MainWindow{
		window:   win,
		appState: appState,
	}

	mw.tabs = container.NewDocTabs()
	mw.tabs.CloseIntercept = func(item *container.TabItem) {
		mw.closeTab(item)
	}

	mw.buildMenu()
	mw.addKeyboardShortcuts()

	mw.showWelcomeTab()
	win.SetContent(mw.tabs)

	return mw
}

func (mw *MainWindow) buildMenu() {
	openItem := fyne.NewMenuItem("Open...", func() {
		mw.showOpenDialog()
	})

	saveFilteredItem := fyne.NewMenuItem("Save Filtered As...", func() {
		mw.showSaveFilteredDialog()
	})

	fileMenu := fyne.NewMenu("File", openItem, saveFilteredItem)
	mw.window.SetMainMenu(fyne.NewMainMenu(fileMenu))
}

func (mw *MainWindow) addKeyboardShortcuts() {
	// Ctrl+O / Cmd+O: Open file
	mw.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierShortcutDefault},
		func(_ fyne.Shortcut) {
			mw.showOpenDialog()
		},
	)

	// Ctrl+C / Cmd+C: Copy selected line
	mw.window.Canvas().AddShortcut(
		&fyne.ShortcutCopy{},
		func(_ fyne.Shortcut) {
			if lt := mw.activeLogTab(); lt != nil {
				lt.CopySelectedLine()
			}
		},
	)

	// Ctrl+W / Cmd+W: Close current tab
	mw.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyW, Modifier: fyne.KeyModifierShortcutDefault},
		func(_ fyne.Shortcut) {
			if item := mw.tabs.Selected(); item != nil {
				mw.closeTab(item)
			}
		},
	)

	// Ctrl+F / Cmd+F: Format selected line as JSON
	mw.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierShortcutDefault},
		func(_ fyne.Shortcut) {
			if lt := mw.activeLogTab(); lt != nil {
				lt.FormatSelectedAsJSON()
			}
		},
	)
}

func (mw *MainWindow) showOpenDialog() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		reader.Close()
		path := reader.URI().Path()
		mw.OpenFile(path)
	}, mw.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".log", ".txt", ".json", ".csv"}))
	fd.Show()
}

func (mw *MainWindow) showSaveFilteredDialog() {
	lt := mw.activeLogTab()
	if lt == nil || lt.filterResult == nil {
		return
	}

	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		for _, lineNum := range lt.filterResult.OriginalLineNumbers {
			line := lt.reader.Line(lineNum)
			io.WriteString(writer, line+"\n")
		}
	}, mw.window)

	fd.SetFileName("filtered-" + filepath.Base(lt.reader.FilePath()))
	fd.Show()
}

// OpenFile opens a file and creates a new tab for it.
func (mw *MainWindow) OpenFile(path string) {
	go func() {
		reader, err := fileio.Open(path, nil)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, mw.window)
			})
			return
		}

		mw.appState.TrackReader(reader)

		fyne.Do(func() {
			mw.removeWelcomeTab()

			lt := NewLogTab(reader, mw.appState, mw.window, func(p string) {
				mw.OpenFile(p)
			})

			mw.logTabs = append(mw.logTabs, lt)

			tabItem := lt.TabItem()
			mw.tabs.Append(tabItem)
			mw.tabs.Select(tabItem)
		})
	}()
}

func (mw *MainWindow) closeTab(item *container.TabItem) {
	for i, lt := range mw.logTabs {
		if lt.tabItem == item {
			mw.logTabs = append(mw.logTabs[:i], mw.logTabs[i+1:]...)
			break
		}
	}
	mw.tabs.Remove(item)
	if len(mw.tabs.Items) == 0 {
		mw.showWelcomeTab()
	}
}

func (mw *MainWindow) activeLogTab() *LogTab {
	selected := mw.tabs.Selected()
	if selected == nil {
		return nil
	}
	for _, lt := range mw.logTabs {
		if lt.tabItem == selected {
			return lt
		}
	}
	return nil
}

// Window returns the underlying fyne.Window.
func (mw *MainWindow) Window() fyne.Window {
	return mw.window
}

// Show displays the main window.
func (mw *MainWindow) Show() {
	mw.window.ShowAndRun()
}

func (mw *MainWindow) showWelcomeTab() {
	shortcut := "Ctrl+O"
	if runtime.GOOS == "darwin" {
		shortcut = "Cmd+O"
	}

	btn := widget.NewButton("Click here or press "+shortcut+" to open a file", func() {
		mw.showOpenDialog()
	})

	content := container.New(layout.NewCenterLayout(), btn)
	mw.welcomeTab = container.NewTabItem("Welcome", content)
	mw.tabs.Append(mw.welcomeTab)
	mw.tabs.Select(mw.welcomeTab)
}

func (mw *MainWindow) removeWelcomeTab() {
	if mw.welcomeTab != nil {
		mw.tabs.Remove(mw.welcomeTab)
		mw.welcomeTab = nil
	}
}
