package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type FixedLayout struct {
	size fyne.Size
}

func (l *FixedLayout) Layout(childs []fyne.CanvasObject, _ fyne.Size) {
	for _, c := range childs {
		c.Move(fyne.NewPos(0, 0))
		c.Resize(l.size)
	}
}

func (l *FixedLayout) MinSize(childs []fyne.CanvasObject) fyne.Size {
	return l.size
}

func NewFixed(size fyne.Size, childs ...fyne.CanvasObject) *fyne.Container {
	return container.New(&FixedLayout{size: size}, childs...)
}

type ListLayout struct{}

func (l *ListLayout) maxMinSize(childs []fyne.CanvasObject) fyne.Size {
	var maxMin fyne.Size
	for _, c := range childs {
		maxMin = maxMin.Max(c.MinSize())
	}
	return maxMin
}

func (l *ListLayout) Layout(childs []fyne.CanvasObject, size fyne.Size) {
	maxMin := l.maxMinSize(childs)
	for y, c := range childs {
		posX := (size.Width - maxMin.Width) / 2
		posY := float32(y) * maxMin.Height
		pos := fyne.NewPos(posX, posY)
		c.Move(pos)
		c.Resize(maxMin)
	}
}

func (l *ListLayout) MinSize(childs []fyne.CanvasObject) fyne.Size {
	maxMin := l.maxMinSize(childs)
	width := maxMin.Width
	height := maxMin.Height * float32(len(childs))
	minSize := fyne.NewSize(width, height)
	return minSize
}

func NewList() *fyne.Container {
	return container.New(&ListLayout{})
}

func NewPanel(content fyne.CanvasObject) fyne.Widget {
	return widget.NewCard("", "", content)
}

func NewFixedPanel(size fyne.Size, content fyne.CanvasObject) *fyne.Container {
	return NewFixed(size, widget.NewCard("", "", content))
}

type TappableIcon struct {
	widget.Icon
	onTapped func()
}

func NewTappableIcon(res fyne.Resource, onTapped func()) *TappableIcon {
	icon := &TappableIcon{}
	icon.ExtendBaseWidget(icon)
	icon.SetResource(res)
	icon.onTapped = onTapped
	return icon
}

func (t *TappableIcon) Tapped(_ *fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func NewCenteredLabel(text string, wrap bool, italic bool, bold bool) *widget.Label {
	label := widget.NewLabel(text)
	label.Alignment = fyne.TextAlignCenter
	if wrap {
		label.Wrapping = fyne.TextWrapWord
	}
	label.TextStyle.Italic = italic
	label.TextStyle.Bold = bold
	return label
}

func NewHelpIcon(message string) fyne.CanvasObject {
	return container.NewCenter(
		NewFixed(fyne.NewSize(20, 20),
			NewTappableIcon(theme.HelpIcon(), func() {
				go ShowInfoPopup("Info", message)
			}),
		),
	)
}
