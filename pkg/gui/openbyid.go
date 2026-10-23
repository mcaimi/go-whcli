package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var openByIDDialogSize = fyne.NewSize(360, 160)

// ShowOpenByIDDialog opens a small dialog to look up a wallpaper directly
// by ID (parity with the CLI's "info" command), reachable from the File menu.
func ShowOpenByIDDialog(w fyne.Window, detailPanel *DetailPanel, setStatus func(string)) {
	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Wallpaper ID")

	form := widget.NewForm(widget.NewFormItem("Wallpaper ID", idEntry))

	var d *dialog.ConfirmDialog
	d = dialog.NewCustomConfirm("Open by ID", "Open", "Cancel", form, func(open bool) {
		if !open || idEntry.Text == "" {
			return
		}
		openWallpaperByID(detailPanel, idEntry.Text, setStatus)
	}, w)
	idEntry.OnSubmitted = func(string) {
		d.Hide()
		if idEntry.Text != "" {
			openWallpaperByID(detailPanel, idEntry.Text, setStatus)
		}
	}
	d.Resize(openByIDDialogSize)
	d.Show()
}
