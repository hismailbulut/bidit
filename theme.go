package main

import (
	"image/color"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type BiditTheme struct{}

func randomColor() color.Color {
	c := [4]byte{}
	rand.Read(c[:])
	return color.RGBA{R: c[0], G: c[1], B: c[2], A: c[3]}
}

func (m BiditTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// switch name {
	// case theme.ColorNameBackground:
	//     return color.NRGBA{0xff, 0xff, 0xff, 0xff}
	// case theme.ColorNameButton:
	//     return color.Transparent
	// case theme.ColorNameDisabled:
	//     return color.NRGBA{0x0, 0x0, 0x0, 0x42}
	// case theme.ColorNameDisabledButton:
	//     return color.NRGBA{0xe5, 0xe5, 0xe5, 0xff}
	// case theme.ColorNameError:
	//     return errorColor
	// case theme.ColorNameForeground:
	//     return color.NRGBA{0x21, 0x21, 0x21, 0xff}
	// case theme.ColorNameHover:
	//     return color.NRGBA{0x0, 0x0, 0x0, 0x0f}
	// case theme.ColorNameInputBackground:
	//     return color.NRGBA{0x0, 0x0, 0x0, 0x19}
	// case theme.ColorNamePlaceHolder:
	//     return color.NRGBA{0x88, 0x88, 0x88, 0xff}
	// case theme.ColorNamePressed:
	//     return color.NRGBA{0x0, 0x0, 0x0, 0x19}
	// case theme.ColorNameScrollBar:
	//     return color.NRGBA{0x0, 0x0, 0x0, 0x99}
	// case theme.ColorNameShadow:
	//     return color.NRGBA{0x0, 0x0, 0x0, 0x33}
	// }
	// return theme.DefaultTheme().Color(name, theme.VariantDark)
	return randomColor()
}

func (m BiditTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m BiditTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m BiditTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
