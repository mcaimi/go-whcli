package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mcaimi/go-whcli/pkg/api"
	"github.com/mcaimi/go-whcli/pkg/common"
)

var filtersPopUpWidth float32 = 280

// SearchResultHandler is invoked (already on the Fyne UI goroutine, via
// fyne.Do) when a search request completes, successfully or not.
type SearchResultHandler func(resp common.SearchResponse, err error)

// NewSearchPanel builds the search query field, with an inline button that
// drops down the "Advanced Filters" (mirroring the CLI's "search" command
// flags, cmd/whcli/cli.go). onSearch receives the outcome of each search;
// setBusy (may be nil) is called with true right before a search starts and
// false once it completes, so the caller can drive an external trigger
// (e.g. a floating action button) instead of an inline submit button. The
// returned func triggers a search using the current field values.
func NewSearchPanel(w fyne.Window, state *AppState, onSearch SearchResultHandler, setBusy func(bool)) (fyne.CanvasObject, func()) {
	defaults := state.Params()

	queryEntry := widget.NewEntry()
	queryEntry.SetPlaceHolder("e.g. +anime +summer")

	resolutionEntry := widget.NewEntry()
	resolutionEntry.SetText(defaults.Resolution)

	purityEntry := widget.NewEntry()
	purityEntry.SetText(defaults.Purity)

	kindEntry := widget.NewEntry()
	kindEntry.SetText(defaults.Kind)

	colorEntry := widget.NewEntry()
	colorEntry.SetPlaceHolder("e.g. 660000")

	orderEntry := widget.NewEntry()
	orderEntry.SetPlaceHolder("asc, desc")

	sortingEntry := widget.NewEntry()
	sortingEntry.SetPlaceHolder("date_added, relevance, random, views, favorites, toplist")

	filters := container.New(layout.NewFormLayout(),
		widget.NewLabel("Resolution"), resolutionEntry,
		widget.NewLabel("Purity"), purityEntry,
		widget.NewLabel("Kind"), kindEntry,
		widget.NewLabel("Color"), colorEntry,
		widget.NewLabel("Order"), orderEntry,
		widget.NewLabel("Sorting"), sortingEntry,
	)
	filtersBox := container.NewPadded(filters)
	filtersBox.Resize(fyne.NewSize(filtersPopUpWidth, filtersBox.MinSize().Height))

	submit := func() {
		params := state.Params()
		params.Query = queryEntry.Text
		params.Resolution = resolutionEntry.Text
		params.Purity = purityEntry.Text
		params.Kind = kindEntry.Text
		params.Color = colorEntry.Text
		params.Order = orderEntry.Text
		params.Sorting = sortingEntry.Text
		params.Page = 1

		if setBusy != nil {
			setBusy(true)
		}
		go func() {
			resp, err := api.Search(params)
			if err == nil {
				params.Seed = fmt.Sprintf("%s", resp.Meta.Seed)
			}
			state.SetParams(params)
			fyne.Do(func() {
				if setBusy != nil {
					setBusy(false)
				}
				onSearch(resp, err)
			})
		}()
	}
	queryEntry.OnSubmitted = func(string) { submit() }

	filtersPopUp := widget.NewPopUp(filtersBox, w.Canvas())
	queryEntry.ActionItem = newTappableIcon(theme.MenuDropDownIcon(), func() {
		if filtersPopUp.Visible() {
			filtersPopUp.Hide()
			return
		}
		filtersPopUp.Resize(fyne.NewSize(filtersPopUpWidth, filtersBox.MinSize().Height))

		// Align the popup's right edge with the window's right edge,
		// directly under the search box, instead of under the chevron.
		entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(queryEntry)
		canvasWidth := w.Canvas().Size().Width
		x := canvasWidth - filtersPopUpWidth
		y := entryPos.Y + queryEntry.Size().Height
		filtersPopUp.ShowAtPosition(fyne.NewPos(x, y))
	})

	return queryEntry, submit
}
