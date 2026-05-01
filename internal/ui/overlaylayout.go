package ui

import "fyne.io/fyne/v2"

// overlayLayout sizes to the first object only; additional objects are
// positioned at (0,0) with the same size but do not affect MinSize.
type overlayLayout struct{}

func (o *overlayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) > 0 {
		return objects[0].MinSize()
	}
	return fyne.NewSize(0, 0)
}

func (o *overlayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		obj.Resize(size)
		obj.Move(fyne.NewPos(0, 0))
	}
}
