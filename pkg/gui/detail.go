package gui

import (
	"fmt"
	"image"
	"path"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mcaimi/go-whcli/pkg/common"
)

var (
	detailPreviewSize = fyne.NewSize(400, 250)
	folderPickerSize  = fyne.NewSize(900, 650)
)

// ShowWallpaperDetail opens a dialog with the wallpaper's filename and
// resolution up front, the rest of its metadata and tags tucked behind a
// collapsible accordion, a larger preview, and Download/Close buttons.
func ShowWallpaperDetail(w fyne.Window, info common.WallpaperInfoResponse, loader *ThumbLoader) {
	data := info.Data

	preview := canvas.NewImageFromResource(theme.FileImageIcon())
	preview.FillMode = canvas.ImageFillContain
	preview.SetMinSize(detailPreviewSize)
	loader.Load(data.Thumbs.Large, func(img image.Image, err error) {
		if err != nil {
			return
		}
		preview.Image = img
		preview.Resource = nil
		preview.Refresh()
	})

	summary := container.New(layout.NewFormLayout(),
		widget.NewLabel("Filename"), widget.NewLabel(path.Base(data.Path)),
		widget.NewLabel("Resolution"), widget.NewLabel(data.Resolution),
	)

	metadata := container.New(layout.NewFormLayout(),
		widget.NewLabel("ID"), widget.NewLabel(data.ID),
		widget.NewLabel("Purity"), widget.NewLabel(data.Purity),
		widget.NewLabel("Category"), widget.NewLabel(data.Category),
		widget.NewLabel("File Size"), widget.NewLabel(fmt.Sprintf("%d bytes", data.FileSize)),
		widget.NewLabel("Uploader"), widget.NewLabel(data.Uploader.Username),
		widget.NewLabel("Views"), widget.NewLabel(fmt.Sprintf("%d", data.Views)),
		widget.NewLabel("Favorites"), widget.NewLabel(fmt.Sprintf("%d", data.Favorites)),
	)

	tagNames := make([]string, 0, len(data.Tags))
	for _, t := range data.Tags {
		tagNames = append(tagNames, t.Name)
	}
	tagsLabel := widget.NewLabel(strings.Join(tagNames, ", "))
	tagsLabel.Wrapping = fyne.TextWrapWord

	metadataAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Metadata", metadata),
		widget.NewAccordionItem("Tags", tagsLabel),
	)

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	statusLabel.Hide()

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	var d *dialog.CustomDialog

	var downloadButton *widget.Button
	downloadButton = widget.NewButton("Download...", func() {
		folderPicker := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Error: %s", err))
				statusLabel.Show()
				return
			}
			if uri == nil {
				return // user cancelled the folder picker
			}
			startDownload(data.ID, uri.Path(), downloadButton, progressBar, statusLabel)
		}, w)
		folderPicker.Resize(folderPickerSize)
		folderPicker.Show()
	})
	closeButton := widget.NewButton("Close", func() { d.Hide() })
	buttons := container.NewHBox(layout.NewSpacer(), downloadButton, closeButton)

	scrollable := container.NewVScroll(container.NewVBox(
		container.NewCenter(preview),
		summary,
		metadataAccordion,
	))

	bottom := container.NewVBox(progressBar, statusLabel, buttons)

	content := container.NewBorder(nil, bottom, nil, nil, scrollable)

	d = dialog.NewCustomWithoutButtons(fmt.Sprintf("Wallpaper %s", data.ID), content, w)
	d.Resize(fyne.NewSize(520, 640))
	d.Show()
}
