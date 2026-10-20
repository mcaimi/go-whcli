package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mcaimi/go-whcli/pkg/api"
)

var settingsDialogSize = fyne.NewSize(480, 360)

const (
	prefAPIKey          = "apiKey"
	prefAPITimeout      = "apiTimeout"
	prefDownloadTimeout = "downloadTimeout"
	prefStartupMode     = "startupMode"
	prefAppTheme        = "appTheme"

	defaultAPITimeout      = "30s"
	defaultDownloadTimeout = "10m"

	// StartupModeLatest and StartupModeToplist are the two choices for what
	// the grid shows on launch, before the user has searched for anything.
	StartupModeLatest  = "Latest"
	StartupModeToplist = "Toplist"

	// ThemeDark and ThemeLight are the two choices for the application's
	// color theme.
	ThemeDark  = "Dark"
	ThemeLight = "Light"
)

// StartupMode returns the persisted startup display mode (StartupModeLatest
// or StartupModeToplist), defaulting to StartupModeLatest.
func StartupMode(a fyne.App) string {
	return a.Preferences().StringWithFallback(prefStartupMode, StartupModeLatest)
}

// AppTheme returns the persisted theme choice (ThemeDark or ThemeLight),
// defaulting to ThemeDark.
func AppTheme(a fyne.App) string {
	return a.Preferences().StringWithFallback(prefAppTheme, ThemeDark)
}

// applyTheme switches the running app's color theme to match the given
// choice (ThemeDark or ThemeLight).
func applyTheme(a fyne.App, choice string) {
	if choice == ThemeLight {
		a.Settings().SetTheme(theme.LightTheme())
		return
	}
	a.Settings().SetTheme(theme.DarkTheme())
}

// ApplySavedSettings reads persisted settings (API key, timeouts) from the
// app's preferences and applies them to the api package, mirroring the
// CLI's --timeout/--download-timeout flags and defaults (cmd/whcli/cli.go).
// A saved API key only overrides WALLHAVEN_API_KEY (already picked up by
// pkg/api's init) when it is non-empty, so an unset Settings field never
// clobbers an env-provided key.
func ApplySavedSettings(a fyne.App) {
	prefs := a.Preferences()

	applyTheme(a, AppTheme(a))

	if key := prefs.String(prefAPIKey); key != "" {
		api.SetApiKey(key)
	}
	if d, err := time.ParseDuration(prefs.StringWithFallback(prefAPITimeout, defaultAPITimeout)); err == nil {
		api.SetApiTimeout(d)
	}
	if d, err := time.ParseDuration(prefs.StringWithFallback(prefDownloadTimeout, defaultDownloadTimeout)); err == nil {
		api.SetDownloadTimeout(d)
	}
}

// ShowSettingsDialog opens a dialog to view/edit the API key and timeouts.
// Saving persists the values to Preferences and applies them immediately.
func ShowSettingsDialog(w fyne.Window, a fyne.App) {
	prefs := a.Preferences()

	keyEntry := widget.NewPasswordEntry()
	keyEntry.SetText(prefs.String(prefAPIKey))

	apiTimeoutEntry := widget.NewEntry()
	apiTimeoutEntry.SetText(prefs.StringWithFallback(prefAPITimeout, defaultAPITimeout))

	downloadTimeoutEntry := widget.NewEntry()
	downloadTimeoutEntry.SetText(prefs.StringWithFallback(prefDownloadTimeout, defaultDownloadTimeout))

	startupModeRadio := widget.NewRadioGroup([]string{StartupModeLatest, StartupModeToplist}, nil)
	startupModeRadio.Horizontal = true
	startupModeRadio.SetSelected(StartupMode(a))

	themeRadio := widget.NewRadioGroup([]string{ThemeDark, ThemeLight}, nil)
	themeRadio.Horizontal = true
	themeRadio.SetSelected(AppTheme(a))

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("API Key", keyEntry),
			widget.NewFormItem("API Timeout", apiTimeoutEntry),
			widget.NewFormItem("Download Timeout", downloadTimeoutEntry),
			widget.NewFormItem("On Launch, Show", startupModeRadio),
			widget.NewFormItem("Theme", themeRadio),
		},
	}

	d := dialog.NewCustomConfirm("Settings", "Save", "Cancel", form, func(save bool) {
		if !save {
			return
		}

		apiTimeout, err := time.ParseDuration(apiTimeoutEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid API timeout %q (use Go duration format, e.g. 30s, 5m): %w", apiTimeoutEntry.Text, err), w)
			return
		}
		downloadTimeout, err := time.ParseDuration(downloadTimeoutEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid download timeout %q (use Go duration format, e.g. 5m, 10m): %w", downloadTimeoutEntry.Text, err), w)
			return
		}

		startupMode := startupModeRadio.Selected
		if startupMode == "" {
			startupMode = StartupModeLatest
		}
		appTheme := themeRadio.Selected
		if appTheme == "" {
			appTheme = ThemeDark
		}

		prefs.SetString(prefAPIKey, keyEntry.Text)
		prefs.SetString(prefAPITimeout, apiTimeoutEntry.Text)
		prefs.SetString(prefDownloadTimeout, downloadTimeoutEntry.Text)
		prefs.SetString(prefStartupMode, startupMode)
		prefs.SetString(prefAppTheme, appTheme)

		if keyEntry.Text != "" {
			api.SetApiKey(keyEntry.Text)
		}
		api.SetApiTimeout(apiTimeout)
		api.SetDownloadTimeout(downloadTimeout)
		applyTheme(a, appTheme)
	}, w)
	d.Resize(settingsDialogSize)
	d.Show()
}
