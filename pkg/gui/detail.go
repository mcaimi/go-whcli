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

// DetailPanel is a persistent side-panel widget showing the currently
// selected wallpaper's preview, metadata, tags, and download controls. It
// is built once and updated in place via Show (instead of opening a new
// dialog per selection), so it can live alongside the results grid in a
// master-detail layout.
type DetailPanel struct {
	window fyne.Window
	loader *ThumbLoader

	card       *widget.Card
	titleLabel *widget.Label
	preview    *canvas.Image

	filenameLabel   *widget.Label
	resolutionLabel *widget.Label

	idLabel        *widget.Label
	purityLabel    *widget.Label
	categoryLabel  *widget.Label
	fileSizeLabel  *widget.Label
	uploaderLabel  *widget.Label
	viewsLabel     *widget.Label
	favoritesLabel *widget.Label
	tagsLabel      *widget.Label

	downloadButton *widget.Button
	progressBar    *widget.ProgressBar
	statusLabel    *widget.Label

	onVisibilityChange func(visible bool)

	currentID string
	// generation is bumped on every Show, so an in-flight download's
	// async callbacks can detect that the panel has since moved on to a
	// different wallpaper and skip stale UI updates.
	generation uint64
}

// NewDetailPanel builds the (initially empty) detail panel.
func NewDetailPanel(w fyne.Window, loader *ThumbLoader) *DetailPanel {
	p := &DetailPanel{window: w, loader: loader}

	p.preview = canvas.NewImageFromResource(theme.FileImageIcon())
	p.preview.FillMode = canvas.ImageFillContain
	p.preview.SetMinSize(detailPreviewSize)

	p.filenameLabel = widget.NewLabel("")
	p.resolutionLabel = widget.NewLabel("")
	summary := container.New(layout.NewFormLayout(),
		widget.NewLabel("Filename"), p.filenameLabel,
		widget.NewLabel("Resolution"), p.resolutionLabel,
	)

	p.idLabel = widget.NewLabel("")
	p.purityLabel = widget.NewLabel("")
	p.categoryLabel = widget.NewLabel("")
	p.fileSizeLabel = widget.NewLabel("")
	p.uploaderLabel = widget.NewLabel("")
	p.viewsLabel = widget.NewLabel("")
	p.favoritesLabel = widget.NewLabel("")
	metadata := container.New(layout.NewFormLayout(),
		widget.NewLabel("ID"), p.idLabel,
		widget.NewLabel("Purity"), p.purityLabel,
		widget.NewLabel("Category"), p.categoryLabel,
		widget.NewLabel("File Size"), p.fileSizeLabel,
		widget.NewLabel("Uploader"), p.uploaderLabel,
		widget.NewLabel("Views"), p.viewsLabel,
		widget.NewLabel("Favorites"), p.favoritesLabel,
	)

	p.tagsLabel = widget.NewLabel("")
	p.tagsLabel.Wrapping = fyne.TextWrapWord

	metadataAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Metadata", metadata),
		widget.NewAccordionItem("Tags", p.tagsLabel),
	)

	p.statusLabel = widget.NewLabel("")
	p.statusLabel.Wrapping = fyne.TextWrapWord
	p.statusLabel.Hide()

	p.progressBar = widget.NewProgressBar()
	p.progressBar.Hide()

	p.downloadButton = widget.NewButton("Download...", func() { p.pickDownloadFolder() })
	p.downloadButton.Disable()

	hideButton := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { p.setVisible(false) })
	hideButton.Importance = widget.LowImportance

	p.titleLabel = widget.NewLabel("No wallpaper selected")
	p.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	header := container.NewBorder(nil, nil, nil, hideButton, p.titleLabel)

	scrollable := container.NewVScroll(container.NewVBox(
		container.NewCenter(p.preview),
		summary,
		metadataAccordion,
	))
	bottom := container.NewVBox(p.progressBar, p.statusLabel, p.downloadButton)
	cardContent := container.NewBorder(header, bottom, nil, nil, scrollable)

	p.card = widget.NewCard("", "", cardContent)
	p.card.Hide()

	return p
}

// Content returns the panel's root CanvasObject for embedding in a layout.
func (p *DetailPanel) Content() fyne.CanvasObject {
	return p.card
}

// SetOnVisibilityChange registers a callback fired whenever the panel
// should be shown or hidden (via Show, or the panel's own hide button), so
// the caller can drive its own slide-in/out animation. The panel itself
// stays unaware of animation details: it only reports intent here.
func (p *DetailPanel) SetOnVisibilityChange(f func(visible bool)) {
	p.onVisibilityChange = f
}

// FinishHide makes the panel Fyne-hidden. Call this once a hide animation
// driven via SetOnVisibilityChange(false) has finished collapsing the
// panel to zero width — not immediately on setVisible(false), so the panel
// stays visible (and animatable) for the whole slide-out.
func (p *DetailPanel) FinishHide() {
	p.card.Hide()
}

func (p *DetailPanel) setVisible(visible bool) {
	if visible {
		p.card.Show()
	}
	if p.onVisibilityChange != nil {
		p.onVisibilityChange(visible)
	}
}

// Show updates the panel in place to display info, replacing whatever was
// shown before.
func (p *DetailPanel) Show(info common.WallpaperInfoResponse) {
	data := info.Data

	p.setVisible(true)

	p.generation++
	gen := p.generation
	p.currentID = data.ID

	p.titleLabel.SetText(fmt.Sprintf("Wallpaper %s", data.ID))

	p.filenameLabel.SetText(path.Base(data.Path))
	p.resolutionLabel.SetText(data.Resolution)

	p.idLabel.SetText(data.ID)
	p.purityLabel.SetText(data.Purity)
	p.categoryLabel.SetText(data.Category)
	p.fileSizeLabel.SetText(fmt.Sprintf("%d bytes", data.FileSize))
	p.uploaderLabel.SetText(data.Uploader.Username)
	p.viewsLabel.SetText(fmt.Sprintf("%d", data.Views))
	p.favoritesLabel.SetText(fmt.Sprintf("%d", data.Favorites))

	tagNames := make([]string, 0, len(data.Tags))
	for _, t := range data.Tags {
		tagNames = append(tagNames, t.Name)
	}
	p.tagsLabel.SetText(strings.Join(tagNames, ", "))

	p.progressBar.Hide()
	p.statusLabel.Hide()
	p.downloadButton.Enable()

	p.preview.Image = nil
	p.preview.Resource = theme.FileImageIcon()
	p.preview.Refresh()
	p.loader.Load(data.Thumbs.Large, func(img image.Image, err error) {
		if gen != p.generation || err != nil {
			return
		}
		p.preview.Image = img
		p.preview.Resource = nil
		p.preview.Refresh()
	})
}

// pickDownloadFolder opens the folder picker for the wallpaper currently
// shown, and starts the download if the user picks a destination.
func (p *DetailPanel) pickDownloadFolder() {
	gen := p.generation
	id := p.currentID
	stillValid := func() bool { return gen == p.generation }

	folderPicker := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if !stillValid() {
			return
		}
		if err != nil {
			p.statusLabel.SetText(fmt.Sprintf("Error: %s", err))
			p.statusLabel.Show()
			return
		}
		if uri == nil {
			return // user cancelled the folder picker
		}
		startDownload(id, uri.Path(), p.downloadButton, p.progressBar, p.statusLabel, stillValid)
	}, p.window)
	folderPicker.Resize(folderPickerSize)
	folderPicker.Show()
}
