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

var filtersPopUpWidth float32 = 380

// kindLabels and purityLabels name the 3 bit positions of the CLI's "Kind"
// and "Purity" bitfield flags (cmd/whcli/cli.go), in order.
var (
	kindLabels   = [3]string{"Anime", "Nature", "People"}
	purityLabels = [3]string{"SFW", "Sketchy", "NSFW"}
)

// bitsToOptions converts a 3-character "xxx" bitfield string (e.g. "111")
// into the labels whose bit is set.
func bitsToOptions(bits string, labels [3]string) []string {
	selected := make([]string, 0, len(labels))
	for i, label := range labels {
		if i < len(bits) && bits[i] == '1' {
			selected = append(selected, label)
		}
	}
	return selected
}

// optionsToBits is the inverse of bitsToOptions: it builds a 3-character
// "xxx" bitfield string with a '1' for each selected label.
func optionsToBits(selected []string, labels [3]string) string {
	bits := [3]byte{'0', '0', '0'}
	for _, s := range selected {
		for i, label := range labels {
			if s == label {
				bits[i] = '1'
			}
		}
	}
	return string(bits[:])
}

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

	purityCheck := widget.NewCheckGroup(purityLabels[:], nil)
	purityCheck.Horizontal = true
	purityCheck.SetSelected(bitsToOptions(defaults.Purity, purityLabels))

	kindCheck := widget.NewCheckGroup(kindLabels[:], nil)
	kindCheck.Horizontal = true
	kindCheck.SetSelected(bitsToOptions(defaults.Kind, kindLabels))

	colorEntry := widget.NewEntry()
	colorEntry.SetPlaceHolder("e.g. 660000")

	orderSelect := widget.NewSelect([]string{"asc", "desc"}, nil)
	orderSelect.PlaceHolder = "(default)"
	if defaults.Order != "" {
		orderSelect.SetSelected(defaults.Order)
	}

	sortingSelect := widget.NewSelect([]string{"date_added", "relevance", "random", "views", "favorites", "toplist"}, nil)
	sortingSelect.PlaceHolder = "(default)"
	if defaults.Sorting != "" {
		sortingSelect.SetSelected(defaults.Sorting)
	}

	filters := container.New(layout.NewFormLayout(),
		widget.NewLabel("Resolution"), resolutionEntry,
		widget.NewLabel("Purity"), purityCheck,
		widget.NewLabel("Kind"), kindCheck,
		widget.NewLabel("Color"), colorEntry,
		widget.NewLabel("Order"), orderSelect,
		widget.NewLabel("Sorting"), sortingSelect,
	)
	filtersBox := container.NewPadded(filters)
	filtersBox.Resize(fyne.NewSize(filtersPopUpWidth, filtersBox.MinSize().Height))

	submit := func() {
		params := state.Params()
		params.Query = queryEntry.Text
		params.Resolution = resolutionEntry.Text
		params.Purity = optionsToBits(purityCheck.Selected, purityLabels)
		params.Kind = optionsToBits(kindCheck.Selected, kindLabels)
		params.Color = colorEntry.Text
		params.Order = orderSelect.Selected
		params.Sorting = sortingSelect.Selected
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
	filtersButton := widget.NewButtonWithIcon("Advanced", theme.MenuDropDownIcon(), func() {})
	filtersButton.Importance = widget.LowImportance
	filtersButton.IconPlacement = widget.ButtonIconTrailingText
	filtersButton.OnTapped = func() {
		if filtersPopUp.Visible() {
			filtersPopUp.Hide()
			return
		}
		filtersPopUp.Resize(fyne.NewSize(filtersPopUpWidth, filtersBox.MinSize().Height))

		// Anchor to the button itself, like widget.PopUpMenu does via
		// ShowAtRelativePosition, rather than computing an absolute
		// position by hand. The X offset keeps the popup's right edge
		// flush with the button's right edge (which sits at the far
		// right of the search bar), so it still reads as right-aligned.
		rel := fyne.NewPos(filtersButton.Size().Width-filtersPopUpWidth, filtersButton.Size().Height)
		filtersPopUp.ShowAtRelativePosition(rel, filtersButton)
	}
	queryEntry.ActionItem = filtersButton

	return queryEntry, submit
}
