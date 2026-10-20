package gui

import (
	"fmt"
	"image"
	"image/color"
	"path"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mcaimi/go-whcli/pkg/api"
	"github.com/mcaimi/go-whcli/pkg/common"
)

var gridCellSize = fyne.NewSize(200, 150)

const (
	selectedShadowBlurRadius = 12
	selectedCardCornerRadius = 8
)

var selectedShadowColor = color.NRGBA{R: 128, G: 128, B: 128, A: 180}

// thumbCard is a tappable thumbnail + filename card used as a grid cell. It
// can be marked as the current selection, drawing a shadow around it.
type thumbCard struct {
	widget.BaseWidget
	bg      *canvas.Rectangle
	content fyne.CanvasObject
	onTap   func()
}

func newThumbCard(inner fyne.CanvasObject, onTap func()) *thumbCard {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = selectedCardCornerRadius
	c := &thumbCard{bg: bg, content: container.NewStack(bg, inner), onTap: onTap}
	c.ExtendBaseWidget(c)
	return c
}

func (c *thumbCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

func (c *thumbCard) Tapped(_ *fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

// SetSelected toggles the selection shadow around the card.
func (c *thumbCard) SetSelected(selected bool) {
	if selected {
		c.bg.Shadow = canvas.Shadow{
			Color:      selectedShadowColor,
			BlurRadius: selectedShadowBlurRadius,
			Variant:    canvas.BoxShadow,
		}
	} else {
		c.bg.Shadow = canvas.Shadow{}
	}
	c.bg.Refresh()
}

// ResultsGrid renders a page of wallpaper search results as a thumbnail
// grid with Prev/Next pagination controls.
type ResultsGrid struct {
	state  *AppState
	loader *ThumbLoader

	onSelect func(item common.SearchItem)
	onError  func(err error)

	grid       *fyne.Container
	pageLabel  *widget.Label
	prevButton *widget.Button
	nextButton *widget.Button
	content    *fyne.Container

	current      common.SearchResponse
	cellTokens   []uint64
	tokenSeq     uint64
	selectedCard *thumbCard
}

// NewResultsGrid builds an empty results grid wired to state for pagination
// requests. onSelect fires (on the UI goroutine) when a thumbnail is
// clicked; onError fires when a pagination request fails.
func NewResultsGrid(state *AppState, loader *ThumbLoader, onSelect func(item common.SearchItem), onError func(err error)) *ResultsGrid {
	g := &ResultsGrid{
		state:    state,
		loader:   loader,
		onSelect: onSelect,
		onError:  onError,
	}

	g.grid = container.New(newCenteredGridWrapLayout(gridCellSize))
	g.pageLabel = widget.NewLabel("No search yet.")

	g.prevButton = widget.NewButton("< Prev", func() { g.gotoPage(g.current.Meta.CurrentPage - 1) })
	g.nextButton = widget.NewButton("Next >", func() { g.gotoPage(g.current.Meta.CurrentPage + 1) })
	g.prevButton.Disable()
	g.nextButton.Disable()

	pagination := container.NewHBox(g.prevButton, g.pageLabel, g.nextButton)
	g.content = container.NewBorder(nil, pagination, nil, nil, container.NewVScroll(g.grid))

	return g
}

// Content returns the grid's root CanvasObject for embedding in a layout.
func (g *ResultsGrid) Content() fyne.CanvasObject {
	return g.content
}

// SetResults replaces the grid with a new page of results and kicks off
// thumbnail loading for each cell.
func (g *ResultsGrid) SetResults(resp common.SearchResponse) {
	g.current = resp
	g.tokenSeq++

	g.grid.RemoveAll()
	g.cellTokens = make([]uint64, len(resp.Data))
	g.selectedCard = nil

	for i, item := range resp.Data {
		idx := i
		token := g.tokenSeq
		g.cellTokens[idx] = token

		thumb := canvas.NewImageFromResource(theme.FileImageIcon())
		thumb.FillMode = canvas.ImageFillContain
		thumb.SetMinSize(gridCellSize)

		label := widget.NewLabel(path.Base(item.Path))
		label.Alignment = fyne.TextAlignCenter
		label.Truncation = fyne.TextTruncateEllipsis

		cellItem := item
		var card *thumbCard
		card = newThumbCard(container.NewBorder(nil, label, nil, nil, thumb), func() {
			if g.selectedCard != nil {
				g.selectedCard.SetSelected(false)
			}
			card.SetSelected(true)
			g.selectedCard = card

			if g.onSelect != nil {
				g.onSelect(cellItem)
			}
		})
		g.grid.Add(card)

		g.loader.Load(item.Thumbs.Small, func(img image.Image, err error) {
			// Stale-request guard: a page change may have reassigned this
			// cell (or rebuilt the grid) before this fetch completed.
			if idx >= len(g.cellTokens) || g.cellTokens[idx] != token {
				return
			}
			if err != nil {
				return
			}
			thumb.Image = img
			thumb.Resource = nil
			thumb.Refresh()
		})
	}

	g.grid.Refresh()

	g.pageLabel.SetText(fmt.Sprintf("Page %d / %d (%d results)", resp.Meta.CurrentPage, resp.Meta.LastPage, resp.Meta.Total))
	g.updatePaginationButtons()
}

func (g *ResultsGrid) updatePaginationButtons() {
	if g.current.Meta.CurrentPage > 1 {
		g.prevButton.Enable()
	} else {
		g.prevButton.Disable()
	}
	if g.current.Meta.LastPage == 0 || g.current.Meta.CurrentPage < g.current.Meta.LastPage {
		g.nextButton.Enable()
	} else {
		g.nextButton.Disable()
	}
}

func (g *ResultsGrid) gotoPage(page int) {
	if page < 1 {
		return
	}
	params := g.state.Params()
	params.Page = page

	g.prevButton.Disable()
	g.nextButton.Disable()
	go func() {
		resp, err := api.Search(params)
		fyne.Do(func() {
			if err != nil {
				g.updatePaginationButtons()
				if g.onError != nil {
					g.onError(err)
				}
				return
			}
			g.state.SetParams(params)
			g.SetResults(resp)
		})
	}()
}
