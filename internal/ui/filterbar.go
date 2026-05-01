package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FilterBar is the filter input area at the bottom of a log tab.
type FilterBar struct {
	Entry         *widget.Entry
	CaseSensitive *widget.Check
	FilterBtn     *widget.Button
	FilterTabBtn  *widget.Button
	ClearBtn      *widget.Button
	MatchCount    *widget.Label
	container     fyne.CanvasObject
}

// NewFilterBar creates a filter bar with the given callbacks.
// onFilter is called when "Filter" is clicked.
// onFilterToTab is called when "Filter to Tab" is clicked.
func NewFilterBar(onFilter func(query string, caseSensitive bool), onFilterToTab func(query string, caseSensitive bool), onClear func()) *FilterBar {
	f := &FilterBar{
		Entry:         widget.NewEntry(),
		CaseSensitive: widget.NewCheck("Case sensitive", nil),
		MatchCount:    widget.NewLabel(""),
	}

	f.Entry.SetPlaceHolder("Text filter...")

	f.FilterBtn = widget.NewButton("Filter", func() {
		if onFilter != nil {
			onFilter(f.Entry.Text, f.CaseSensitive.Checked)
		}
	})

	f.FilterTabBtn = widget.NewButton("Filter to Tab", func() {
		if onFilterToTab != nil {
			onFilterToTab(f.Entry.Text, f.CaseSensitive.Checked)
		}
	})

	f.ClearBtn = widget.NewButton("Clear", func() {
		f.Entry.SetText("")
		if onClear != nil {
			onClear()
		}
	})

	// Submit on Enter key in the entry
	f.Entry.OnSubmitted = func(s string) {
		if onFilter != nil {
			onFilter(s, f.CaseSensitive.Checked)
		}
	}

	return f
}

// Container returns the filter bar as a renderable container.
func (f *FilterBar) Container() fyne.CanvasObject {
	return container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewLabel("Text filter:"),
		),
		container.NewHBox(
			f.CaseSensitive,
			f.FilterBtn,
			f.FilterTabBtn,
			f.ClearBtn,
			f.MatchCount,
		),
		f.Entry,
	)
}

// SetMatchCount updates the match count label.
func (f *FilterBar) SetMatchCount(count int) {
	if count < 0 {
		f.MatchCount.SetText("")
	} else {
		f.MatchCount.SetText(fmt.Sprintf("%d matches", count))
	}
}
