package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const fabDiameter = 56

// fabButton is a round, icon-only floating action button.
type fabButton struct {
	widget.DisableableWidget
	bg    *canvas.Circle
	icon  *widget.Icon
	onTap func()
}

func newFabButton(res fyne.Resource, onTap func()) *fabButton {
	b := &fabButton{
		bg:    canvas.NewCircle(theme.Color(theme.ColorNamePrimary)),
		icon:  widget.NewIcon(res),
		onTap: onTap,
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *fabButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(b.bg, container.NewCenter(b.icon)))
}

func (b *fabButton) MinSize() fyne.Size {
	return fyne.NewSize(fabDiameter, fabDiameter)
}

func (b *fabButton) Tapped(_ *fyne.PointEvent) {
	if b.Disabled() {
		return
	}
	if b.onTap != nil {
		b.onTap()
	}
}

// Refresh keeps the background dimmed while the button is disabled.
func (b *fabButton) Refresh() {
	if b.Disabled() {
		b.bg.FillColor = theme.Color(theme.ColorNameDisabled)
	} else {
		b.bg.FillColor = theme.Color(theme.ColorNamePrimary)
	}
	b.bg.Refresh()
	b.BaseWidget.Refresh()
}
