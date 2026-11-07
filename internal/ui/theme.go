package ui

import (
	"image/color"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type customTheme struct {
	isMobile bool
}

// scrollTrackAlpha makes the track behind a widened scroll bar see-through. The bar widens over
// the content while the pointer is on it, and Fyne's opaque track would hide what is underneath.
const scrollTrackAlpha = 0x4d // 30 %

func (m customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	c := theme.DefaultTheme().Color(name, variant)
	if name == theme.ColorNameScrollBarBackground {
		n := color.NRGBAModel.Convert(c).(color.NRGBA)
		n.A = scrollTrackAlpha
		return n
	}
	return c
}

func (m customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m customTheme) Size(name fyne.ThemeSizeName) float32 {
	// SizeNameInlineIcon keeps the Fyne default (20) everywhere: buttons are at least
	// this tall, so a larger value makes them taller than entries in the same row

	switch name {
	case theme.SizeNameCardRadius:
		return 8
	case widgets.SizeNameCardPadding:
		if m.isMobile {
			return 12
		}
		return 10
	case widgets.SizeNameDisplayText:
		return 56
	}

	// Adjust sizes for mobile
	if m.isMobile {
		switch name {
		case theme.SizeNamePadding:
			return 8 // Larger padding for better touch targets in menus
		case theme.SizeNameInnerPadding:
			return 6 // Larger inner padding for menus
		case theme.SizeNameText:
			return 16 // Larger text for menus
		case theme.SizeNameScrollBarSmall:
			return 2
		}
	}

	return theme.DefaultTheme().Size(name)
}

func NewCustomTheme() fyne.Theme {
	return customTheme{isMobile: fyne.CurrentDevice().IsMobile()}
}
