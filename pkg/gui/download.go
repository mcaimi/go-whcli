package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/mcaimi/go-whcli/pkg/api"
	"github.com/mcaimi/go-whcli/pkg/common"
)

// progressUpdateInterval throttles UI progress-bar updates: the download
// core reports progress once per common.CHUNK_SIZE (1KB) chunk, which would
// otherwise flood fyne.Do with thousands of calls for a multi-MB wallpaper.
const progressUpdateInterval = 100 * time.Millisecond

// startDownload runs the wallpaper download in the background, updating
// progressBar/statusLabel on the UI goroutine, and disables downloadButton
// while the transfer is in flight.
func startDownload(id, destDir string, downloadButton *widget.Button, progressBar *widget.ProgressBar, statusLabel *widget.Label) {
	downloadButton.Disable()
	progressBar.SetValue(0)
	progressBar.Show()
	statusLabel.SetText("Starting download...")
	statusLabel.Show()

	params := common.QueryParameters{Id: id, Path: destDir}

	var lastUpdate time.Time
	go func() {
		destFile, err := api.DownloadWallpaperWithProgress(params,
			func(_ string, transferred, total int64) {
				if time.Since(lastUpdate) < progressUpdateInterval {
					return
				}
				lastUpdate = time.Now()
				if total <= 0 {
					return
				}
				value := float64(transferred) / float64(total)
				fyne.Do(func() { progressBar.SetValue(value) })
			},
			func(msg string) {
				if msg == "" {
					return
				}
				fyne.Do(func() { statusLabel.SetText(msg) })
			},
		)
		fyne.Do(func() {
			downloadButton.Enable()
			if err != nil {
				progressBar.Hide()
				statusLabel.SetText(fmt.Sprintf("Download failed: %s", api.ApiErrorMessage("download of wallpaper", id, err)))
				return
			}
			progressBar.SetValue(1)
			statusLabel.SetText(fmt.Sprintf("Saved to %s", destFile))
		})
	}()
}
