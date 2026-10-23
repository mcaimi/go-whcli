package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mcaimi/go-whcli/pkg/api"
	"github.com/mcaimi/go-whcli/pkg/common"
)

const appID = "io.github.mcaimi.whcli-gui"

// Run builds and shows the main application window, then blocks until it closes.
func Run() {
	a := app.NewWithID(appID)
	ApplySavedSettings(a)

	w := a.NewWindow("Wallpapers (powered by wallhaven.cc)")
	w.Resize(fyne.NewSize(1000, 700))

	state := NewAppState()
	thumbLoader := NewThumbLoader(NewThumbCache())
	detailPanel := NewDetailPanel(w, thumbLoader)

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	statusLabel.Hide()

	// setStatus shows msg in the status line, or hides the line entirely
	// when there's nothing worth telling the user (e.g. a plain successful
	// search, whose result count is already shown by resultsGrid's own
	// pagination bar).
	setStatus := func(msg string) {
		if msg == "" {
			statusLabel.Hide()
			return
		}
		statusLabel.SetText(msg)
		statusLabel.Show()
	}

	resultsGrid := NewResultsGrid(
		state, thumbLoader,
		func(item common.SearchItem) {
			openWallpaperByID(detailPanel, item.ID, setStatus)
		},
		func(err error) {
			setStatus(fmt.Sprintf("Pagination failed: %s", api.ApiErrorMessage("search page", "", err)))
		},
	)

	var searchFAB *fabButton
	searchPanel, triggerSearch := NewSearchPanel(
		w, state,
		func(resp common.SearchResponse, err error) {
			if err != nil {
				setStatus(fmt.Sprintf("Search failed: %s", api.ApiErrorMessage("search", "", err)))
				return
			}
			// Result count/page info is shown by resultsGrid's own pagination
			// bar, so there's no separate status message for a successful search.
			setStatus("")
			resultsGrid.SetResults(resp)
		},
		func(busy bool) {
			if searchFAB == nil {
				return
			}
			if busy {
				searchFAB.Disable()
			} else {
				searchFAB.Enable()
			}
		},
	)
	searchFAB = newFabButton(theme.SearchIcon(), triggerSearch)

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu(
			"File",
			fyne.NewMenuItem("Open by ID...", func() { ShowOpenByIDDialog(w, detailPanel, setStatus) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Settings...", func() { ShowSettingsDialog(w, a) }),
		),
	))

	top := container.NewVBox(searchPanel, statusLabel)
	gridArea := container.NewBorder(top, nil, nil, nil, resultsGrid.Content())

	randomFAB := NewRandomFAB(state, detailPanel, setStatus)
	// Anchor the floating buttons to the grid area only, so they sit flush
	// with the divider against the detail panel instead of floating over it.
	gridWithFABs := newBottomRightOverlay(gridArea, randomFAB, searchFAB)

	splitLayout := &detailSplitLayout{}
	content := container.New(splitLayout, gridWithFABs, detailPanel.Content())

	var currentAnim *fyne.Animation
	detailPanel.SetOnVisibilityChange(func(visible bool) {
		if currentAnim != nil {
			currentAnim.Stop()
		}
		start := splitLayout.shownFraction
		stop := float32(0)
		if visible {
			stop = 1
			splitLayout.panelWidth = detailPanel.Content().MinSize().Width
		}
		currentAnim = canvas.NewPositionAnimation(
			fyne.NewPos(start, 0), fyne.NewPos(stop, 0),
			canvas.DurationStandard,
			func(pos fyne.Position) {
				splitLayout.shownFraction = pos.X
				content.Refresh()
				if !visible && pos.X <= 0.001 {
					detailPanel.FinishHide()
				}
			},
		)
		currentAnim.Start()
	})
	w.SetContent(content)

	loadStartupWallpapers(a, state, resultsGrid, setStatus)

	w.ShowAndRun()
}

// loadStartupWallpapers populates the grid on startup with either
// Wallhaven's latest or toplist results (per the persisted Settings
// choice), so the window never opens to an empty grid before the user has
// searched for anything.
func loadStartupWallpapers(a fyne.App, state *AppState, grid *ResultsGrid, setStatus func(string)) {
	mode := StartupMode(a)
	label := "Latest Entries"

	params := state.Params()
	params.Page = 1
	if mode == StartupModeToplist {
		params.Sorting = "toplist"
		label = "Top List Entries"
	}

	go func() {
		resp, err := api.Search(params)
		fyne.Do(func() {
			if err != nil {
				setStatus(fmt.Sprintf("Failed to load %s: %s", label, api.ApiErrorMessage("search", "", err)))
				return
			}
			grid.SetResults(resp)
			setStatus(fmt.Sprintf("No search yet - Displaying %s", label))
		})
	}()
}

// openWallpaperByID fetches full wallpaper info by ID and shows it in
// detailPanel on success, or reports the failure via setStatus.
func openWallpaperByID(detailPanel *DetailPanel, id string, setStatus func(string)) {
	setStatus(fmt.Sprintf("Loading wallpaper %s...", id))
	go func() {
		info, err := api.GetById(common.QueryParameters{Id: id})
		fyne.Do(func() {
			if err != nil {
				setStatus(fmt.Sprintf("Lookup failed: %s", api.ApiErrorMessage("wallpaper", id, err)))
				return
			}
			setStatus("")
			detailPanel.Show(info)
		})
	}()
}
