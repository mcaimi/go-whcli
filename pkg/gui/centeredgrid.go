package gui

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// centeredGridWrapLayout behaves like Fyne's built-in GridWrapLayout (fixed
// cell size, dynamic column count based on available width) but horizontally
// centers each full row of cells instead of hugging the left edge, so the
// thumbnail grid doesn't look stranded in a wide window.
type centeredGridWrapLayout struct {
	cellSize fyne.Size
	colCount int
	rowCount int
}

func newCenteredGridWrapLayout(cellSize fyne.Size) fyne.Layout {
	return &centeredGridWrapLayout{cellSize: cellSize, colCount: 1, rowCount: 1}
}

func (g *centeredGridWrapLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	padding := theme.Padding()

	g.colCount = 1
	if size.Width > g.cellSize.Width {
		g.colCount = int(math.Floor(float64(size.Width+padding) / float64(g.cellSize.Width+padding)))
	}
	if g.colCount < 1 {
		g.colCount = 1
	}

	rowWidth := float32(g.colCount)*g.cellSize.Width + float32(g.colCount-1)*padding
	offsetX := (size.Width - rowWidth) / 2
	if offsetX < 0 {
		offsetX = 0
	}

	g.rowCount = 0
	x, y := offsetX, float32(0)
	i := 0
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if i%g.colCount == 0 {
			g.rowCount++
		}

		child.Move(fyne.NewPos(x, y))
		child.Resize(g.cellSize)

		if (i+1)%g.colCount == 0 {
			x = offsetX
			y += g.cellSize.Height + padding
		} else {
			x += g.cellSize.Width + padding
		}
		i++
	}
}

func (g *centeredGridWrapLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	rows := g.rowCount
	if rows < 1 {
		rows = 1
	}
	return fyne.NewSize(g.cellSize.Width, (g.cellSize.Height*float32(rows))+(float32(rows-1)*theme.Padding()))
}
