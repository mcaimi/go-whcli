package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"github.com/mcaimi/go-whcli/pkg/api"
	"github.com/mcaimi/go-whcli/pkg/common"
)

// NewRandomFAB builds a floating round button that fetches a random
// wallpaper using the current filter values in state (mirroring the CLI's
// "random" command) and shows it in the detail panel on success. Unlike
// the CLI, it does not auto-download: the user explicitly clicks Download
// from the detail panel.
func NewRandomFAB(state *AppState, detailPanel *DetailPanel, setStatus func(string)) fyne.CanvasObject {
	var button *fabButton
	button = newFabButton(theme.ViewRefreshIcon(), func() {
		params := state.Params()
		params.Page = 1

		button.Disable()
		setStatus("Fetching a random wallpaper...")
		go func() {
			id, err := api.RandomWallpaper(params)
			if err != nil {
				fyne.Do(func() {
					button.Enable()
					setStatus(fmt.Sprintf("Random wallpaper failed: %s", api.ApiErrorMessage("random wallpaper search", "", err)))
				})
				return
			}
			info, err := api.GetById(common.QueryParameters{Id: id})
			fyne.Do(func() {
				button.Enable()
				if err != nil {
					setStatus(fmt.Sprintf("Random wallpaper failed: %s", api.ApiErrorMessage("wallpaper", id, err)))
					return
				}
				setStatus("")
				detailPanel.Show(info)
			})
		}()
	})
	return button
}
