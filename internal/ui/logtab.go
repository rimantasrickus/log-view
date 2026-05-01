package ui

import (
	"context"
	"fmt"
	"image/color"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"log-viewer/internal/app"
	"log-viewer/internal/fileio"
	"log-viewer/internal/filter"
	"log-viewer/internal/plugin"
)

const maxDisplayLen = 300

// lineNumWidth calculates the width needed for line numbers.
func lineNumWidth(lineCount int) float32 {
	digits := len(strconv.Itoa(lineCount))
	return float32(digits*8 + 16)
}

// LogTab represents a single opened file tab with its list, filter bar, and state.
type LogTab struct {
	reader      *fileio.FileReader
	appState    *app.App
	list        *widget.List
	filterBar   *FilterBar
	statusBar   *StatusBar
	selectedRow int
	content     fyne.CanvasObject
	tabItem     *container.TabItem
	window      fyne.Window
	lnWidth     float32

	// Context menu
	contextMenu  *fyne.Container
	menuOverlay  *fyne.Container
	listWithMenu *fyne.Container
	menuShownAt  time.Time

	// Filter result state
	filterResult  *filter.Result
	filterList    *widget.List
	filterSplit   *container.Split
	mainContainer fyne.CanvasObject
	bottomBar     fyne.CanvasObject
	cancelFilter  context.CancelFunc

	// Callback to open a new tab (injected by window)
	OnOpenTab func(path string)
}

// NewLogTab creates a new log tab for the given file reader.
func NewLogTab(reader *fileio.FileReader, appState *app.App, win fyne.Window, onOpenTab func(string)) *LogTab {
	lt := &LogTab{
		reader:      reader,
		appState:    appState,
		selectedRow: -1,
		window:      win,
		OnOpenTab:   onOpenTab,
	}

	lt.buildList()
	lt.buildFilterBar()
	lt.statusBar = NewStatusBar()
	lt.statusBar.Update(reader.LineCount(), reader.FileSize(), -1)
	lt.bottomBar = container.NewVBox(lt.filterBar.Container(), lt.statusBar.Container())
	lt.buildLayout()

	return lt
}

func (lt *LogTab) buildList() {
	lineCount := lt.reader.LineCount()
	lt.lnWidth = lineNumWidth(1)
	lt.list = widget.NewList(
		func() int {
			return lineCount + 1
		},
		func() fyne.CanvasObject {
			return newListRow(&lt.lnWidth, func(rowID int, pos fyne.Position) {
				lt.ShowContextMenu(pos)
			}, func(rowID int) {
				if rowID < lineCount {
					lt.list.Select(rowID)
				}
			})
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if lt.contextMenu != nil && time.Since(lt.menuShownAt) > 150*time.Millisecond {
				lt.hideContextMenu()
			}
			row := obj.(*listRow)
			row.rowID = id

			if id >= lineCount {
				row.lineNum.Text = ""
				row.lineNum.Refresh()
				row.text.Text = ""
				row.text.Refresh()
				row.fullLine = ""
				return
			}

			// Dynamically widen the line number column when needed
			needed := lineNumWidth(id + 1)
			if needed > lt.lnWidth {
				lt.lnWidth = needed
			}

			row.lineNum.Text = strconv.Itoa(id + 1)
			row.lineNum.Refresh()

			line := lt.reader.Line(id)
			row.fullLine = line
			if len(line) > maxDisplayLen {
				line = line[:maxDisplayLen]
			}
			row.text.Text = line
			row.text.Refresh()
		},
	)

	lt.list.OnSelected = func(id widget.ListItemID) {
		if id >= lineCount {
			return
		}
		lt.selectedRow = id
		lt.statusBar.Update(lt.reader.LineCount(), lt.reader.FileSize(), id)
		lt.hideContextMenu()
	}
}

func (lt *LogTab) buildFilterBar() {
	lt.filterBar = NewFilterBar(
		func(query string, caseSensitive bool) {
			lt.runFilter(query, caseSensitive, false)
		},
		func(query string, caseSensitive bool) {
			lt.runFilter(query, caseSensitive, true)
		},
		func() {
			lt.clearFilter()
		},
	)
}

func (lt *LogTab) buildLayout() {
	lt.menuOverlay = container.NewWithoutLayout()
	lt.listWithMenu = container.New(&overlayLayout{}, lt.list, lt.menuOverlay)
	lt.mainContainer = container.NewBorder(nil, lt.bottomBar, nil, nil, lt.listWithMenu)
	lt.content = lt.mainContainer
}

func (lt *LogTab) runFilter(query string, caseSensitive bool, toNewTab bool) {
	if query == "" {
		lt.clearFilter()
		return
	}

	if lt.cancelFilter != nil {
		lt.cancelFilter()
	}

	ctx, cancel := context.WithCancel(context.Background())
	lt.cancelFilter = cancel

	lt.filterBar.SetMatchCount(-1)

	go func() {
		result, err := filter.Run(ctx, lt.reader, query, caseSensitive, nil)
		if err != nil {
			return
		}

		lines := make([]string, len(result.OriginalLineNumbers))
		for i, lineNum := range result.OriginalLineNumbers {
			lines[i] = fmt.Sprintf("[%d] %s", lineNum+1, lt.reader.Line(lineNum))
		}

		if toNewTab {
			tmpPath, err := lt.appState.TempMgr.CreateTempFile(lines)
			if err != nil {
				return
			}
			fyne.Do(func() {
				lt.filterResult = result
				if lt.OnOpenTab != nil {
					lt.OnOpenTab(tmpPath)
				}
				lt.filterBar.SetMatchCount(len(result.OriginalLineNumbers))
			})
		} else {
			fyne.Do(func() {
				lt.filterResult = result
				lt.showFilterResults(result)
				lt.filterBar.SetMatchCount(len(result.OriginalLineNumbers))
			})
		}
	}()
}

func (lt *LogTab) showFilterResults(result *filter.Result) {
	lt.filterList = widget.NewList(
		func() int {
			return len(result.OriginalLineNumbers)
		},
		func() fyne.CanvasObject {
			lineNum := canvas.NewText("0", nil)
			lineNum.TextStyle = fyne.TextStyle{Monospace: true}
			lineNum.TextSize = 13

			text := canvas.NewText("", nil)
			text.TextStyle = fyne.TextStyle{Monospace: true}
			text.TextSize = 13

			numBox := container.NewWithoutLayout(lineNum)
			numBox.Resize(fyne.NewSize(lt.lnWidth, 20))

			return container.NewBorder(nil, nil, numBox, nil, text)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			border := obj.(*fyne.Container)
			numBox := border.Objects[1].(*fyne.Container)
			lineNum := numBox.Objects[0].(*canvas.Text)
			text := border.Objects[0].(*canvas.Text)

			lineIdx := result.OriginalLineNumbers[id]
			lineNum.Text = strconv.Itoa(lineIdx + 1)
			lineNum.Refresh()

			line := lt.reader.Line(lineIdx)
			if len(line) > maxDisplayLen {
				line = line[:maxDisplayLen]
			}
			text.Text = line
			text.Refresh()
		},
	)

	lt.filterList.OnSelected = func(id widget.ListItemID) {
		lineIdx := result.OriginalLineNumbers[id]
		lt.list.ScrollTo(lineIdx)
		lt.list.Select(lineIdx)
	}

	// Layout: list on top, then filterBar + filterList + statusBar at bottom
	bottomPanel := container.NewBorder(
		lt.filterBar.Container(),
		lt.statusBar.Container(),
		nil, nil,
		lt.filterList,
	)

	lt.filterSplit = container.NewVSplit(lt.listWithMenu, bottomPanel)
	lt.filterSplit.SetOffset(0.6)

	lt.content = lt.filterSplit

	if lt.tabItem != nil {
		lt.tabItem.Content = lt.content
	}
}

func (lt *LogTab) clearFilter() {
	lt.filterResult = nil
	lt.filterList = nil
	lt.filterBar.SetMatchCount(-1)
	lt.buildLayout()
	if lt.tabItem != nil {
		lt.tabItem.Content = lt.content
	}
}

// Content returns the tab content for embedding in a DocTabs.
func (lt *LogTab) Content() fyne.CanvasObject {
	return lt.content
}

// TabItem returns a DocTabs tab item for this log tab.
func (lt *LogTab) TabItem() *container.TabItem {
	name := filepath.Base(lt.reader.FilePath())
	lt.tabItem = container.NewTabItem(name, lt.content)
	return lt.tabItem
}

// SelectedLine returns the full text of the currently selected line, or empty string.
func (lt *LogTab) SelectedLine() string {
	if lt.selectedRow < 0 || lt.selectedRow >= lt.reader.LineCount() {
		return ""
	}
	return lt.reader.Line(lt.selectedRow)
}

// CopySelectedLine copies the currently selected line to the clipboard.
func (lt *LogTab) CopySelectedLine() {
	line := lt.SelectedLine()
	if line == "" {
		return
	}
	lt.window.Clipboard().SetContent(line)
}

// ShowContextMenu shows a right-click context menu at the given position.
func (lt *LogTab) ShowContextMenu(pos fyne.Position) {
	lt.hideContextMenu()

	line := lt.SelectedLine()

	var buttons []fyne.CanvasObject

	copyBtn := widget.NewButton("Copy Line", func() {
		lt.hideContextMenu()
		lt.CopySelectedLine()
	})
	copyBtn.Importance = widget.LowImportance
	copyBtn.Alignment = widget.ButtonAlignLeading
	buttons = append(buttons, copyBtn)

	if line != "" {
		handlers := plugin.FindHandlers(line)
		for _, h := range handlers {
			handler := h
			btn := widget.NewButton(handler.Name(), func() {
				lt.hideContextMenu()
				lt.showPluginPopup(handler, line)
			})
			btn.Importance = widget.LowImportance
			btn.Alignment = widget.ButtonAlignLeading
			buttons = append(buttons, btn)
		}
	}

	// Convert absolute position to overlay-relative
	overlayAbs := fyne.CurrentApp().Driver().AbsolutePositionForObject(lt.menuOverlay)
	localPos := fyne.NewPos(pos.X-overlayAbs.X, pos.Y-overlayAbs.Y)

	menuBox := container.NewVBox(buttons...)
	menuSize := menuBox.MinSize()
	w := max(menuSize.Width, 160)
	h := menuSize.Height

	th := fyne.CurrentApp().Settings().Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	bgColor := th.Color(theme.ColorNameMenuBackground, v)
	sepColor := th.Color(theme.ColorNameSeparator, v)
	if bgColor == (color.Color)(nil) {
		bgColor = color.NRGBA{R: 40, G: 40, B: 40, A: 245}
	}

	border := canvas.NewRectangle(sepColor)
	border.CornerRadius = 6
	border.Resize(fyne.NewSize(w+2, h+2))
	border.Move(localPos.Add(fyne.NewDelta(-1, -1)))

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 5
	bg.Resize(fyne.NewSize(w, h))
	bg.Move(localPos)

	menuBox.Resize(fyne.NewSize(w, h))
	menuBox.Move(localPos)

	lt.contextMenu = container.NewWithoutLayout(border, bg, menuBox)
	lt.menuShownAt = time.Now()
	lt.menuOverlay.Objects = []fyne.CanvasObject{lt.contextMenu}
	lt.menuOverlay.Refresh()
}

func (lt *LogTab) hideContextMenu() {
	if lt.contextMenu != nil {
		lt.menuOverlay.Objects = nil
		lt.menuOverlay.Refresh()
		lt.contextMenu = nil
	}
}

func (lt *LogTab) showPluginPopup(p plugin.LinePlugin, line string) {
	rendered := p.Render(line)

	popupWin := fyne.CurrentApp().NewWindow(p.Name())
	popupWin.SetContent(container.NewPadded(rendered))
	popupWin.Resize(fyne.NewSize(800, 600))
	popupWin.Show()
}

// Reader returns the file reader for this tab.
func (lt *LogTab) Reader() *fileio.FileReader {
	return lt.reader
}

// FormatSelectedAsJSON opens Format as JSON window for the selected line if applicable.
func (lt *LogTab) FormatSelectedAsJSON() {
	line := lt.SelectedLine()
	if line == "" {
		return
	}
	handlers := plugin.FindHandlers(line)
	for _, h := range handlers {
		if h.Name() == "Format as JSON" {
			lt.showPluginPopup(h, line)
			return
		}
	}
}
