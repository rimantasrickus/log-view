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
	case theme.ColorNameHeaderBackground:
		// Slightly darker header to distinguish tab bar from content
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xe8, G: 0xe8, B: 0xe8, A: 0xff}
		}
		return color.NRGBA{R: 0x25, G: 0x25, B: 0x25, A: 0xff}
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
