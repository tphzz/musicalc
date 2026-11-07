package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// TestScrollTrackIsSeeThrough checks that the widened scroll bar's track keeps Fyne's colour but
// lets the content under it show, in both variants, and that other colours are unchanged
func TestScrollTrackIsSeeThrough(t *testing.T) {
	th := customTheme{}
	for _, v := range []fyne.ThemeVariant{theme.VariantDark, theme.VariantLight} {
		got := color.NRGBAModel.Convert(th.Color(theme.ColorNameScrollBarBackground, v)).(color.NRGBA)
		want := color.NRGBAModel.Convert(theme.DefaultTheme().Color(theme.ColorNameScrollBarBackground, v)).(color.NRGBA)
		if got.A != scrollTrackAlpha || got.R != want.R || got.G != want.G || got.B != want.B {
			t.Errorf("variant %v: track %v, want %v at alpha %#x", v, got, want, scrollTrackAlpha)
		}
		for _, name := range []fyne.ThemeColorName{theme.ColorNameScrollBar, theme.ColorNameBackground, theme.ColorNameForeground} {
			if th.Color(name, v) != theme.DefaultTheme().Color(name, v) {
				t.Errorf("variant %v: %s changed", v, name)
			}
		}
	}
}
