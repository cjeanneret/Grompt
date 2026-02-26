package ui

import (
	"errors"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"grompt/internal/content"
	"grompt/internal/input"
	scrollengine "grompt/internal/scroll"
)

const (
	appName        = "grompt"
	defaultWidth   = 1024
	defaultHeight  = 768
	initialMessage = "Open an HTML or Markdown file to start."
)

func Run() error {
	a := app.NewWithID("com.grompt.app")
	typographyTheme := NewTypographyTheme(DefaultContentFontSize)
	a.Settings().SetTheme(typographyTheme)

	w := a.NewWindow(appName)
	w.Resize(fyne.NewSize(defaultWidth, defaultHeight))

	initialContent := widget.NewRichTextFromMarkdown(initialMessage)
	initialContent.Wrapping = fyne.TextWrapWord
	content.ApplyTypography(initialContent)

	scroll := container.NewScroll(initialContent)
	scroll.SetMinSize(fyne.NewSize(640, 400))

	var engine *scrollengine.Engine
	engine = scrollengine.NewEngine(func(delta float64) {
		fyne.Do(func() {
			if scroll.Content == nil {
				return
			}

			maxOffset := scroll.Content.MinSize().Height - scroll.Size().Height
			if maxOffset <= 0 {
				engine.Pause()
				return
			}

			nextOffset := scroll.Offset.Y + float32(delta)
			if nextOffset >= maxOffset {
				nextOffset = maxOffset
				engine.Pause()
			}
			scroll.ScrollToOffset(fyne.NewPos(0, nextOffset))
		})
	})
	defer engine.Stop()

	var controls *Controls
	controls = NewControls(ControlActions{
		OnOpen: func() {
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
				scroll.ScrollToOffset(fyne.Position{})
				scroll.Refresh()
				controls.SetFileName(filepath.Base(path))
			}, w)

			fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown", ".html", ".htm"}))
			fileDialog.Show()
		},
		OnPlay: func() {
			engine.Play()
		},
		OnPause: func() {
			engine.Pause()
		},
		OnSpeedUp: func() {
			controls.SetSpeed(engine.SpeedUp())
		},
		OnSpeedDown: func() {
			controls.SetSpeed(engine.SpeedDown())
		},
		OnFontSizeUp: func() {
			controls.SetFontSize(typographyTheme.IncreaseBodySize())
			a.Settings().SetTheme(typographyTheme)
			if scroll.Content != nil {
				scroll.Content.Refresh()
			}
			scroll.Refresh()
		},
		OnFontSizeDown: func() {
			controls.SetFontSize(typographyTheme.DecreaseBodySize())
			a.Settings().SetTheme(typographyTheme)
			if scroll.Content != nil {
				scroll.Content.Refresh()
			}
			scroll.Refresh()
		},
	}, engine.Speed(), typographyTheme.BodySize())

	input.BindTeleprompterKeys(w.Canvas(), input.KeyActions{
		OnTogglePlayPause: func() {
			engine.Toggle()
		},
		OnSpeedUp: func() {
			controls.SetSpeed(engine.SpeedUp())
		},
		OnSpeedDown: func() {
			controls.SetSpeed(engine.SpeedDown())
		},
	})

	w.SetContent(container.NewBorder(controls.View(), nil, nil, nil, scroll))
	w.ShowAndRun()
	return nil
}
