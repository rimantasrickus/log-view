package jsonformat

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	jsonTextSize = 14
)

type jsonTextNode struct {
	key      string
	value    any
	nodeType string // object, array, value
	children []*jsonTextNode
	expanded bool
}

type jsonTextLine struct {
	text     string
	node     *jsonTextNode
	foldable bool
}

type selectionPoint struct {
	line int
	col  int
}

type selectionRange struct {
	start selectionPoint
	end   selectionPoint
}

func (s selectionRange) valid() bool {
	return s.start.line >= 0 && s.end.line >= 0 && (s.start.line != s.end.line || s.start.col != s.end.col)
}

func (s selectionRange) normalized() (selectionPoint, selectionPoint) {
	if s.start.line < s.end.line || (s.start.line == s.end.line && s.start.col <= s.end.col) {
		return s.start, s.end
	}
	return s.end, s.start
}

func (s selectionRange) contains(line int) bool {
	if !s.valid() {
		return false
	}
	start, end := s.normalized()
	return line >= start.line && line <= end.line
}

func (s selectionRange) lineRange(line int) (int, int, bool) {
	if !s.valid() {
		return 0, 0, false
	}
	start, end := s.normalized()
	if line < start.line || line > end.line {
		return 0, 0, false
	}
	if start.line == end.line {
		return start.col, end.col, true
	}
	if line == start.line {
		return start.col, -1, true
	}
	if line == end.line {
		return 0, end.col, true
	}
	return 0, -1, true
}

func (s selectionRange) reset() selectionRange {
	return selectionRange{start: selectionPoint{line: -1, col: 0}, end: selectionPoint{line: -1, col: 0}}
}

// JSONTextView is a custom widget that renders collapsible JSON as plain
// text and supports text selection and copy.
type JSONTextView struct {
	widget.BaseWidget
	root       *jsonTextNode
	lines      []*jsonTextLine
	selection  selectionRange
	charWidth  float32
	lineHeight float32
}

func NewJSONTextView(data any) *JSONTextView {
	v := &JSONTextView{}
	v.ExtendBaseWidget(v)
	v.selection = v.selection.reset()
	v.charWidth = fyne.MeasureText("M", jsonTextSize, fyne.TextStyle{Monospace: true}).Width
	v.lineHeight = fyne.MeasureText("M", jsonTextSize, fyne.TextStyle{Monospace: true}).Height + 4
	v.SetJSON(data)
	return v
}

func (v *JSONTextView) SetJSON(data any) {
	v.root = v.buildNode("", data)
	v.lines = v.buildLines()
	v.selection = v.selection.reset()
	v.Refresh()
}

func (v *JSONTextView) SelectedText() string {
	if !v.selection.valid() {
		return ""
	}

	start, end := v.selection.normalized()
	if start.line >= len(v.lines) || end.line >= len(v.lines) {
		return ""
	}

	var sb strings.Builder
	for line := start.line; line <= end.line; line++ {
		row := v.lines[line].text
		runes := []rune(row)
		if line == start.line && line == end.line {
			if start.col < 0 {
				start.col = 0
			}
			if end.col > len(runes) {
				end.col = len(runes)
			}
			sb.WriteString(string(runes[start.col:end.col]))
			break
		}
		if line == start.line {
			if start.col < 0 {
				start.col = 0
			}
			sb.WriteString(string(runes[start.col:]))
			sb.WriteString("\n")
			continue
		}
		if line == end.line {
			if end.col > len(runes) {
				end.col = len(runes)
			}
			sb.WriteString(string(runes[:end.col]))
			break
		}
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (v *JSONTextView) CopySelection() {
	text := v.SelectedText()
	if text == "" {
		return
	}
	fyne.CurrentApp().Clipboard().SetContent(text)
}

func (v *JSONTextView) CreateRenderer() fyne.WidgetRenderer {
	return newJSONTextViewRenderer(v)
}

func (v *JSONTextView) TypedShortcut(shortcut fyne.Shortcut) {
	switch shortcut.(type) {
	case *fyne.ShortcutCopy:
		v.CopySelection()
	}
}

func (v *JSONTextView) Tapped(e *fyne.PointEvent) {
	line, col := v.pointToLocation(e.Position)
	if line < 0 || line >= len(v.lines) {
		return
	}
	if v.hitFoldMarker(e.Position, line) {
		v.toggleFold(line)
		return
	}
	v.selection.start = selectionPoint{line: line, col: col}
	v.selection.end = v.selection.start
	v.Refresh()
}

func (v *JSONTextView) Dragged(d *fyne.DragEvent) {
	line, col := v.pointToLocation(d.Position)
	if line < 0 || line >= len(v.lines) {
		return
	}
	v.selection.end = selectionPoint{line: line, col: col}
	v.Refresh()
}

func (v *JSONTextView) DragEnd() {}

func (v *JSONTextView) hitFoldMarker(pos fyne.Position, line int) bool {
	if line < 0 || line >= len(v.lines) {
		return false
	}
	row := v.lines[line]
	if !row.foldable {
		return false
	}
	return pos.X <= v.charWidth*5.5
}

func (v *JSONTextView) pointToLocation(pos fyne.Position) (int, int) {
	if len(v.lines) == 0 {
		return -1, 0
	}

	line := int(pos.Y / v.lineHeight)
	if line < 0 {
		line = 0
	}
	if line >= len(v.lines) {
		line = len(v.lines) - 1
	}
	col := int(pos.X / v.charWidth)
	if col < 0 {
		col = 0
	}
	runes := []rune(v.lines[line].text)
	if col > len(runes) {
		col = len(runes)
	}
	return line, col
}

func (v *JSONTextView) toggleFold(line int) {
	if line < 0 || line >= len(v.lines) {
		return
	}
	node := v.lines[line].node
	if node == nil || !node.isBranch() {
		return
	}
	node.expanded = !node.expanded
	v.lines = v.buildLines()
	v.Refresh()
}

func (v *JSONTextView) buildLines() []*jsonTextLine {
	if v.root == nil {
		return nil
	}
	var lines []*jsonTextLine
	v.appendLines(&lines, v.root, 0, true)
	return lines
}

func (v *JSONTextView) appendLines(lines *[]*jsonTextLine, node *jsonTextNode, depth int, isRoot bool) {
	prefix := ""
	foldable := node.isBranch()
	if foldable {
		if node.expanded {
			prefix = "[-] "
		} else {
			prefix = "[+] "
		}
	}
	indent := strings.Repeat("  ", depth)
	lineText := prefix + indent

	switch node.nodeType {
	case "object":
		if isRoot && node.key == "" {
			lineText += "{"
		} else {
			lineText += fmt.Sprintf("%q: {", node.key)
		}
	case "array":
		if isRoot && node.key == "" {
			lineText += "["
		} else {
			lineText += fmt.Sprintf("%q: [", node.key)
		}
	case "value":
		if node.key != "" {
			lineText += fmt.Sprintf("%q: %s", node.key, formatValue(node.value))
		} else {
			lineText += formatValue(node.value)
		}
	}

	*lines = append(*lines, &jsonTextLine{text: lineText, node: node, foldable: foldable})

	if node.isBranch() && node.expanded {
		for _, child := range node.children {
			v.appendLines(lines, child, depth+1, false)
		}
		closeText := strings.Repeat("  ", depth)
		if node.nodeType == "object" {
			closeText += "}"
		} else {
			closeText += "]"
		}
		*lines = append(*lines, &jsonTextLine{text: closeText})
	}
}

func (v *JSONTextView) buildNode(key string, value any) *jsonTextNode {
	node := &jsonTextNode{key: key, value: value, expanded: true}
	switch val := value.(type) {
	case map[string]any:
		node.nodeType = "object"
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			node.children = append(node.children, v.buildNode(k, val[k]))
		}
	case []any:
		node.nodeType = "array"
		for i, item := range val {
			childKey := fmt.Sprintf("[%d]", i)
			node.children = append(node.children, v.buildNode(childKey, item))
		}
	default:
		node.nodeType = "value"
	}
	return node
}

func (n *jsonTextNode) isBranch() bool {
	return n.nodeType == "object" || n.nodeType == "array"
}

type jsonTextViewRenderer struct {
	view    *JSONTextView
	objects []fyne.CanvasObject
}

func newJSONTextViewRenderer(view *JSONTextView) *jsonTextViewRenderer {
	return &jsonTextViewRenderer{view: view}
}

func (r *jsonTextViewRenderer) Destroy() {}

func (r *jsonTextViewRenderer) Layout(size fyne.Size) {
	y := float32(0)
	index := 0
	for i := range r.view.lines {
		// Check if this line has a selection background rectangle
		if r.view.selection.valid() && r.view.selection.contains(i) {
			if index < len(r.objects) {
				if rect, ok := r.objects[index].(*canvas.Rectangle); ok {
					rect.Move(fyne.NewPos(0, y))
					rect.Resize(fyne.NewSize(size.Width, r.view.lineHeight))
					index++
				}
			}
		}
		// Position text object
		if index < len(r.objects) {
			if text, ok := r.objects[index].(*canvas.Text); ok {
				text.Move(fyne.NewPos(0, y))
				text.Resize(fyne.NewSize(size.Width, r.view.lineHeight))
				index++
			}
		}
		y += r.view.lineHeight
	}
}

func (r *jsonTextViewRenderer) MinSize() fyne.Size {
	width := float32(0)
	for _, line := range r.view.lines {
		size := fyne.MeasureText(line.text, jsonTextSize, fyne.TextStyle{Monospace: true})
		if size.Width > width {
			width = size.Width
		}
	}
	height := float32(len(r.view.lines)) * r.view.lineHeight
	if width < 100 {
		width = 100
	}
	return fyne.NewSize(width+20, height)
}

func (r *jsonTextViewRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *jsonTextViewRenderer) Refresh() {
	r.objects = make([]fyne.CanvasObject, 0, len(r.view.lines)*2)
	selected := r.view.selection.valid()
	for i, line := range r.view.lines {
		if selected && r.view.selection.contains(i) {
			rect := canvas.NewRectangle(theme.ForegroundColor())
			rect.FillColor = theme.ForegroundColor()
			rect.StrokeWidth = 0
			r.objects = append(r.objects, rect)
		}
		text := canvas.NewText(line.text, theme.ForegroundColor())
		text.TextStyle = fyne.TextStyle{Monospace: true}
		text.TextSize = jsonTextSize
		text.Alignment = fyne.TextAlignLeading
		if selected && r.view.selection.contains(i) {
			text.Color = theme.BackgroundColor()
		} else {
			text.Color = theme.ColorForWidget(theme.ColorNameForeground, r.view)
		}
		r.objects = append(r.objects, text)
	}
}
