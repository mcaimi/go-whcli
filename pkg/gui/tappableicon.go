package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// tappableIcon is a small icon-only widget that invokes onTap when clicked,
// suitable for use as an Entry.ActionItem or other compact inline control.
type tappableIcon struct {
	widget.BaseWidget
	image *canvas.Image
	onTap func()
}

func newTappableIcon(res fyne.Resource, onTap func()) *tappableIcon {
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillContain
	iconSize := theme.IconInlineSize()
	img.SetMinSize(fyne.NewSize(iconSize, iconSize))
	t := &tappableIcon{image: img, onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableIcon) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.image)
}

func (t *tappableIcon) Tapped(_ *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}
