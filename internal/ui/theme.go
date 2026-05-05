package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// logViewerTheme provides custom theming to improve tab visibility.
type logViewerTheme struct{}

var _ fyne.Theme = (*logViewerTheme)(nil)

func NewLogViewerTheme() fyne.Theme {
	return &logViewerTheme{}
}

func (t *logViewerTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameSeparator:
		return theme.DefaultTheme().Color(name, variant)
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *logViewerTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *logViewerTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *logViewerTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
