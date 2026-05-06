package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type CloseableDocTabs struct {
	widget.BaseWidget

	Items []*container.TabItem

	CreateTab      func() *container.TabItem
	CloseIntercept func(*container.TabItem)
	OnClosed       func(*container.TabItem)
	OnSelected     func(*container.TabItem)
	OnUnselected   func(*container.TabItem)

	current  int
	location container.TabLocation

	tabsBar *fyne.Container
	content *fyne.Container
	scroll  *container.Scroll
}

func NewCloseableDocTabs(items ...*container.TabItem) *CloseableDocTabs {
	tabsBar := container.NewHBox()
	scroll := container.NewScroll(tabsBar)
	content := container.NewMax()

	current := -1
	if len(items) > 0 {
		current = 0
	}

	c := &CloseableDocTabs{
		Items:    items,
		current:  current,
		location: container.TabLocationTop,
		tabsBar:  tabsBar,
		content:  content,
		scroll:   scroll,
	}
	c.ExtendBaseWidget(c)
	c.updateTabs()
	return c
}

func (c *CloseableDocTabs) Append(item *container.TabItem) {
	c.Items = append(c.Items, item)
	if len(c.Items) == 1 {
		c.current = 0
	}
	c.updateTabs()
}

func (c *CloseableDocTabs) Remove(item *container.TabItem) {
	for i, it := range c.Items {
		if it == item {
			c.RemoveIndex(i)
			return
		}
	}
}

func (c *CloseableDocTabs) RemoveIndex(index int) {
	if index < 0 || index >= len(c.Items) {
		return
	}

	c.Items = append(c.Items[:index], c.Items[index+1:]...)

	if len(c.Items) == 0 {
		c.current = -1
	} else if c.current == index {
		if index >= len(c.Items) {
			c.current = len(c.Items) - 1
		} else {
			c.current = index
		}
	} else if index < c.current {
		c.current--
	}

	c.updateTabs()
}

func (c *CloseableDocTabs) Select(item *container.TabItem) {
	for i, it := range c.Items {
		if it == item {
			c.SelectIndex(i)
			return
		}
	}
}

func (c *CloseableDocTabs) SelectIndex(index int) {
	if index < 0 || index >= len(c.Items) || c.current == index {
		return
	}

	old := c.Selected()
	if old != nil && c.OnUnselected != nil {
		c.OnUnselected(old)
	}

	c.current = index
	c.updateTabs()

	if selected := c.Selected(); selected != nil && c.OnSelected != nil {
		c.OnSelected(selected)
	}
}

func (c *CloseableDocTabs) Selected() *container.TabItem {
	if c.current < 0 || c.current >= len(c.Items) {
		return nil
	}
	return c.Items[c.current]
}

func (c *CloseableDocTabs) SelectedIndex() int {
	return c.current
}

func (c *CloseableDocTabs) SetItems(items []*container.TabItem) {
	c.Items = items
	if c.current >= len(items) {
		c.current = len(items) - 1
	}
	if c.current < 0 && len(items) > 0 {
		c.current = 0
	}
	c.updateTabs()
}

func (c *CloseableDocTabs) SetTabLocation(l container.TabLocation) {
	c.location = l
	c.updateTabs()
}

func (c *CloseableDocTabs) CreateRenderer() fyne.WidgetRenderer {
	return &closeableDocTabsRenderer{
		docTabs:    c,
		background: canvas.NewRectangle(color.Transparent),
		divider:    canvas.NewRectangle(color.Transparent),
	}
}

func (c *CloseableDocTabs) Show() {
	c.BaseWidget.Show()
}

func (c *CloseableDocTabs) Hide() {
	c.BaseWidget.Hide()
}

func (c *CloseableDocTabs) close(item *container.TabItem) {
	if f := c.CloseIntercept; f != nil {
		f(item)
		return
	}

	c.Remove(item)
	if f := c.OnClosed; f != nil {
		f(item)
	}
}

func (c *CloseableDocTabs) items() []*container.TabItem {
	return c.Items
}

func (c *CloseableDocTabs) selected() int {
	return c.current
}

func (c *CloseableDocTabs) tabLocation() container.TabLocation {
	return c.location
}

func (c *CloseableDocTabs) updateTabs() {
	c.tabsBar.Objects = nil
	for _, item := range c.Items {
		tab := newTabHeader(item.Text, c.Selected() == item,
			func() { c.Select(item) },
			func() { c.close(item) },
			c,
		)
		c.tabsBar.Add(tab)
	}

	c.tabsBar.Refresh()
	c.updateContent()
	c.Refresh()
}

func (c *CloseableDocTabs) updateContent() {
	c.content.Objects = nil
	if selected := c.Selected(); selected != nil && selected.Content != nil {
		c.content.Add(selected.Content)
	}
	c.content.Refresh()
}

type tabHeader struct {
	widget.BaseWidget

	text     string
	selected bool
	onTapped func()
	onClosed func()
	close    *tabCloseButton
	hovered  bool
	tabs     baseTabs
}

type baseTabs interface {
	selected() int
	items() []*container.TabItem
	tabLocation() container.TabLocation
}

func newTabHeader(text string, selected bool, onTapped, onClosed func(), tabs baseTabs) *tabHeader {
	tab := &tabHeader{
		text:     text,
		selected: selected,
		onTapped: onTapped,
		onClosed: onClosed,
		tabs:     tabs,
	}
	tab.close = newTabCloseButton(func() {
		if tab.onClosed != nil {
			tab.onClosed()
		}
	}, tab)
	tab.ExtendBaseWidget(tab)
	return tab
}

func (t *tabHeader) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	indicator := canvas.NewRectangle(color.Transparent)
	th := t.Theme()
	label := canvas.NewText(t.text, nil)
	label.TextSize = th.Size(theme.SizeNameText)
	label.TextStyle.Bold = true
	label.Alignment = fyne.TextAlignCenter

	objects := []fyne.CanvasObject{background, indicator, label, t.close}
	return &tabHeaderRenderer{tab: t, background: background, indicator: indicator, label: label, objects: objects}
}

func (t *tabHeader) MinSize() fyne.Size {
	th := t.Theme()
	closeSize := t.close.MinSize()
	padding := th.Size(theme.SizeNamePadding) * 2

	// Estimate text size
	textSize := th.Size(theme.SizeNameText)
	labelWidth := float32(len(t.text)) * textSize * 0.6 // rough estimate
	labelHeight := textSize * 1.4

	width := labelWidth + closeSize.Width + padding*4
	height := fyne.Max(labelHeight, closeSize.Height) + padding*2
	if height < 40 {
		height = 40
	}
	return fyne.NewSize(width, height)
}

func (t *tabHeader) Tapped(*fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func (t *tabHeader) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonTertiary && t.onClosed != nil {
		t.onClosed()
	}
}

func (t *tabHeader) MouseUp(*desktop.MouseEvent) {}

func (t *tabHeader) MouseIn(*desktop.MouseEvent) {
	t.hovered = true
	t.Refresh()
}

func (t *tabHeader) MouseOut() {
	t.hovered = false
	t.Refresh()
}

func (t *tabHeader) Refresh() {
	t.BaseWidget.Refresh()
}

type tabHeaderRenderer struct {
	tab        *tabHeader
	background *canvas.Rectangle
	indicator  *canvas.Rectangle
	label      *canvas.Text
	objects    []fyne.CanvasObject
}

func (r *tabHeaderRenderer) Destroy() {}

func (r *tabHeaderRenderer) Layout(size fyne.Size) {
	th := r.tab.Theme()
	padding := th.Size(theme.SizeNamePadding) * 2

	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	// Position indicator bar at the bottom
	indicatorHeight := float32(2)
	r.indicator.Resize(fyne.NewSize(size.Width, indicatorHeight))
	r.indicator.Move(fyne.NewPos(0, size.Height-indicatorHeight))

	labelSize := r.label.MinSize()
	closeSize := r.tab.close.MinSize()

	// Center the text, accounting for close button on the right
	textAreaWidth := size.Width - closeSize.Width - padding*2
	if textAreaWidth < labelSize.Width {
		labelSize.Width = textAreaWidth
	}

	r.label.Resize(labelSize)
	r.label.Move(fyne.NewPos(padding, (size.Height-labelSize.Height)/2))

	buttonX := size.Width - closeSize.Width - padding
	r.tab.close.Resize(closeSize)
	r.tab.close.Move(fyne.NewPos(buttonX, (size.Height-closeSize.Height)/2))
}

func (r *tabHeaderRenderer) MinSize() fyne.Size {
	th := r.tab.Theme()
	closeSize := r.tab.close.MinSize()
	padding := th.Size(theme.SizeNamePadding)

	// Estimate text size
	textSize := th.Size(theme.SizeNameText)
	labelWidth := float32(len(r.tab.text)) * textSize * 0.6 // rough estimate
	labelHeight := textSize * 1.2

	width := labelWidth + closeSize.Width + padding*3
	height := fyne.Max(labelHeight, closeSize.Height) + padding*2
	return fyne.NewSize(width, height)
}

func (r *tabHeaderRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *tabHeaderRenderer) Refresh() {
	th := r.tab.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	hovered := r.tab.hovered || r.tab.close.hovered

	// Background is transparent by default
	r.background.FillColor = color.Transparent
	r.background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	r.background.Show()

	// Indicator bar for selected tab
	if r.tab.selected {
		r.indicator.FillColor = th.Color(theme.ColorNamePrimary, v)
		r.indicator.Show()
	} else {
		r.indicator.Hide()
	}

	// Subtle hover rectangle
	if hovered && !r.tab.selected {
		r.background.FillColor = th.Color(theme.ColorNameHover, v)
		r.background.Show()
	}

	if r.tab.selected {
		r.label.Color = th.Color(theme.ColorNamePrimary, v)
		r.label.TextStyle.Bold = true
	} else {
		r.label.Color = th.Color(theme.ColorNameForeground, v)
		r.label.TextStyle.Bold = false
	}

	if r.tab.selected || hovered {
		r.tab.close.Show()
	} else {
		r.tab.close.Hide()
	}

	r.label.Text = r.tab.text
	r.label.Refresh()
	r.tab.close.Refresh()
	canvas.Refresh(r.tab)
}

type tabCloseButton struct {
	widget.BaseWidget
	onTapped     func()
	hovered      bool
	parentHeader *tabHeader
}

func newTabCloseButton(onTapped func(), parent *tabHeader) *tabCloseButton {
	button := &tabCloseButton{onTapped: onTapped, parentHeader: parent}
	button.ExtendBaseWidget(button)
	return button
}

func (b *tabCloseButton) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	icon := canvas.NewImageFromResource(theme.CancelIcon())
	icon.FillMode = canvas.ImageFillContain
	return &tabCloseButtonRenderer{button: b, background: background, icon: icon, objects: []fyne.CanvasObject{background, icon}}
}

func (b *tabCloseButton) MinSize() fyne.Size {
	iconSize := b.Theme().Size(theme.SizeNameInlineIcon)
	return fyne.NewSize(iconSize, iconSize)
}

func (b *tabCloseButton) Tapped(*fyne.PointEvent) {
	if b.onTapped != nil {
		b.onTapped()
	}
}

func (b *tabCloseButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
	if b.parentHeader != nil {
		b.parentHeader.Refresh()
	}
}

func (b *tabCloseButton) MouseMoved(*desktop.MouseEvent) {}

func (b *tabCloseButton) MouseOut() {
	b.hovered = false
	b.Refresh()
	if b.parentHeader != nil {
		b.parentHeader.Refresh()
	}
}

type tabCloseButtonRenderer struct {
	button     *tabCloseButton
	background *canvas.Rectangle
	icon       *canvas.Image
	objects    []fyne.CanvasObject
}

func (r *tabCloseButtonRenderer) Destroy() {}

func (r *tabCloseButtonRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.icon.Resize(size)
}

func (r *tabCloseButtonRenderer) MinSize() fyne.Size {
	return r.button.MinSize()
}

func (r *tabCloseButtonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *tabCloseButtonRenderer) Refresh() {
	th := r.button.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	if r.button.hovered {
		r.background.FillColor = th.Color(theme.ColorNameHover, v)
		r.background.Show()
	} else {
		r.background.FillColor = color.Transparent
		r.background.Hide()
	}
	r.background.Refresh()
	r.icon.Refresh()
}

type closeableDocTabsRenderer struct {
	docTabs    *CloseableDocTabs
	background *canvas.Rectangle
	divider    *canvas.Rectangle
}

func (r *closeableDocTabsRenderer) Destroy() {}

func (r *closeableDocTabsRenderer) Layout(size fyne.Size) {
	scroll := r.docTabs.scroll
	content := r.docTabs.content

	scrollHeight := r.docTabs.tabsBar.MinSize().Height
	contentHeight := size.Height - scrollHeight

	r.background.Resize(fyne.NewSize(size.Width, scrollHeight))
	r.background.Move(fyne.NewPos(0, 0))

	scroll.Resize(fyne.NewSize(size.Width, scrollHeight))
	scroll.Move(fyne.NewPos(0, 0))

	r.divider.Resize(fyne.NewSize(size.Width, 1))
	r.divider.Move(fyne.NewPos(0, scrollHeight-1))

	content.Resize(fyne.NewSize(size.Width, contentHeight))
	content.Move(fyne.NewPos(0, scrollHeight))
}

func (r *closeableDocTabsRenderer) MinSize() fyne.Size {
	scrollMin := r.docTabs.tabsBar.MinSize()
	contentMin := r.docTabs.content.MinSize()
	return fyne.NewSize(fyne.Max(scrollMin.Width, contentMin.Width), scrollMin.Height+contentMin.Height)
}

func (r *closeableDocTabsRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.docTabs.scroll, r.divider, r.docTabs.content}
}

func (r *closeableDocTabsRenderer) Refresh() {
	th := r.docTabs.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.background.FillColor = th.Color(theme.ColorNameHeaderBackground, v)
	r.divider.FillColor = th.Color(theme.ColorNameSeparator, v)
	r.background.Refresh()
	r.divider.Refresh()

	canvas.Refresh(r.docTabs)
}
