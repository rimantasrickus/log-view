package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// newHBox creates a horizontal box with spacing between items.
func newHBox(items ...fyne.CanvasObject) fyne.CanvasObject {
	padded := make([]fyne.CanvasObject, 0, len(items)*2)
	for i, item := range items {
		padded = append(padded, item)
		if i < len(items)-1 {
			padded = append(padded, widget.NewSeparator())
		}
	}
	return container.New(layout.NewHBoxLayout(), padded...)
}
