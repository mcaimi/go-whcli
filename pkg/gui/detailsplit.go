package gui

import "fyne.io/fyne/v2"

// detailSplitLayout lays out exactly two objects: a grid area and a detail
// panel. shownFraction (0..1, animated) controls how much of panelWidth is
// currently reserved for the panel; the grid always gets the remainder, so
// the boundary between them — and anything anchored to the grid's edge,
// like the floating Search/Random buttons — tracks the animation
// continuously instead of only snapping at rest.
type detailSplitLayout struct {
	panelWidth    float32 // natural width of the panel when fully shown
	shownFraction float32 // 0 = fully hidden, 1 = fully shown
}

func (l *detailSplitLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	grid, panel := objects[0], objects[1]

	panelW := l.panelWidth * l.shownFraction
	gridW := size.Width - panelW

	grid.Resize(fyne.NewSize(gridW, size.Height))
	grid.Move(fyne.NewPos(0, 0))

	panel.Resize(fyne.NewSize(panelW, size.Height))
	panel.Move(fyne.NewPos(gridW, 0))
}

func (l *detailSplitLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}
