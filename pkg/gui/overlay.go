package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

const (
	overlayMargin  = 16
	overlaySpacing = 12
)

// bottomRightOverlayLayout lays out the first object as a full-size base
// and stacks any additional objects (e.g. floating action buttons)
// vertically in the bottom-right corner, closest-to-the-corner first, each
// sized to its own MinSize with a fixed margin/spacing between them.
type bottomRightOverlayLayout struct{}

func (bottomRightOverlayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)

	y := size.Height - overlayMargin
	for _, overlay := range objects[1:] {
		overlaySize := overlay.MinSize()
		y -= overlaySize.Height
		overlay.Resize(overlaySize)
		overlay.Move(fyne.NewPos(size.Width-overlaySize.Width-overlayMargin, y))
		y -= overlaySpacing
	}
}

func (bottomRightOverlayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}

// newBottomRightOverlay stacks overlays on top of base, pinned to the
// bottom-right corner of whatever space base is given. overlays are
// stacked closest-to-the-corner first, i.e. overlays[0] sits lowest.
func newBottomRightOverlay(base fyne.CanvasObject, overlays ...fyne.CanvasObject) *fyne.Container {
	objects := append([]fyne.CanvasObject{base}, overlays...)
	return container.New(bottomRightOverlayLayout{}, objects...)
}
