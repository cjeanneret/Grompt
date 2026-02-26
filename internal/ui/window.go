package ui

import (
	"errors"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"grompt/internal/content"
)

const (
	appName        = "grompt"
	defaultWidth   = 1024
	defaultHeight  = 768
	initialMessage = "Open an HTML or Markdown file to start."
)

func Run() error {
	a := app.NewWithID("com.grompt.app")
	w := a.NewWindow(appName)
	w.Resize(fyne.NewSize(defaultWidth, defaultHeight))

	initialContent := widget.NewRichTextFromMarkdown(initialMessage)
	initialContent.Wrapping = fyne.TextWrapWord

	scroll := container.NewScroll(initialContent)
	scroll.SetMinSize(fyne.NewSize(640, 400))
	statusLabel := widget.NewLabel("No file loaded")

	openButton := widget.NewButton("Open", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			path := reader.URI().Path()
			data, format, loadErr := content.LoadFromPath(path)
			if loadErr != nil {
				if errors.Is(loadErr, content.ErrUnsupportedFileType) {
					dialog.ShowInformation("Unsupported file", "Supported extensions are .md, .markdown, .html and .htm.", w)
					return
				}
				dialog.ShowError(loadErr, w)
				return
			}

			rendered, renderErr := content.Render(data, format)
			if renderErr != nil {
				dialog.ShowError(renderErr, w)
				return
			}

			scroll.Content = rendered
			scroll.Offset = fyne.Position{}
			scroll.Refresh()
			statusLabel.SetText(filepath.Base(path))
		}, w)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown", ".html", ".htm"}))
		fileDialog.Show()
	})
	playButton := widget.NewButton("Play", nil)
	pauseButton := widget.NewButton("Pause", nil)
	speedUpButton := widget.NewButton("Speed +", nil)
	speedDownButton := widget.NewButton("Speed -", nil)

	controls := container.NewHBox(
		openButton,
		layout.NewSpacer(),
		statusLabel,
		layout.NewSpacer(),
		playButton,
		pauseButton,
		speedUpButton,
		speedDownButton,
	)

	w.SetContent(container.NewBorder(controls, nil, nil, nil, scroll))
	w.ShowAndRun()
	return nil
}
