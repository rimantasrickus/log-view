package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const rowHeight float32 = 20

// listRow is a custom widget for each row in the log list.
// It implements SecondaryTappable (right-click).
type listRow struct {
	widget.BaseWidget
	lineNum  *canvas.Text
	text     *canvas.Text
	rowID    int
	fullLine string
	lnWidth  *float32

	onRightClick func(rowID int, pos fyne.Position)
	onSelect     func(rowID int)
}

func newListRow(lnWidth *float32, onRightClick func(int, fyne.Position), onSelect func(int)) *listRow {
	lineNum := canvas.NewText("0", nil)
	lineNum.TextStyle = fyne.TextStyle{Monospace: true}
	lineNum.TextSize = 13

	text := canvas.NewText("", nil)
	text.TextStyle = fyne.TextStyle{Monospace: true}
	text.TextSize = 13

	row := &listRow{
		lineNum:      lineNum,
		text:         text,
		lnWidth:      lnWidth,
		onRightClick: onRightClick,
		onSelect:     onSelect,
	}
	row.ExtendBaseWidget(row)
	return row
}

func (r *listRow) CreateRenderer() fyne.WidgetRenderer {
	return &listRowRenderer{row: r}
}

func (r *listRow) Tapped(e *fyne.PointEvent) {
	if r.onSelect != nil {
		r.onSelect(r.rowID)
	}
}

func (r *listRow) TappedSecondary(e *fyne.PointEvent) {
	if r.onSelect != nil {
		r.onSelect(r.rowID)
	}
	if r.onRightClick != nil {
		r.onRightClick(r.rowID, e.AbsolutePosition)
	}
}

type listRowRenderer struct {
	row *listRow
}

func (rr *listRowRenderer) Destroy() {}

func (rr *listRowRenderer) Layout(size fyne.Size) {
	w := *rr.row.lnWidth
	rr.row.lineNum.Move(fyne.NewPos(4, 0))
	rr.row.lineNum.Resize(fyne.NewSize(w-8, size.Height))
	rr.row.text.Move(fyne.NewPos(w, 0))
	rr.row.text.Resize(fyne.NewSize(size.Width-w, size.Height))
}

func (rr *listRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, rowHeight)
}

func (rr *listRowRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{rr.row.lineNum, rr.row.text}
}

func (rr *listRowRenderer) Refresh() {
	rr.row.lineNum.Refresh()
	rr.row.text.Refresh()
}
