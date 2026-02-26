package ui

import (
	"errors"
	"fmt"
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
	wordSpacingMin = 1
	wordSpacingMax = 8
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
	scrollWithFade := NewScrollWithFade(scroll, func() float32 {
		return estimatedLineHeight(typographyTheme.BodySize())
	})

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
	wordSpacing := 1
	var loadedData []byte
	var loadedFormat content.Format
	var loadedFileName string

	refreshViewport := func() {
		a.Settings().SetTheme(typographyTheme)
		if scroll.Content != nil {
			scroll.Content.Refresh()
		}
		scroll.Refresh()
		scrollWithFade.Refresh()
	}

	renderCurrentDocument := func() error {
		if len(loadedData) == 0 {
			return nil
		}

		rendered, renderErr := content.RenderWithOptions(loadedData, loadedFormat, content.RenderOptions{
			WordSpacing: wordSpacing,
		})
		if renderErr != nil {
			return renderErr
		}

		scroll.Content = rendered
		scroll.ScrollToOffset(fyne.Position{})
		refreshViewport()
		controls.SetFileName(loadedFileName)
		return nil
	}

	openFile := func() {
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

			loadedData = data
			loadedFormat = format
			loadedFileName = filepath.Base(path)

			if renderErr := renderCurrentDocument(); renderErr != nil {
				dialog.ShowError(renderErr, w)
			}
		}, w)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown", ".html", ".htm"}))
		fileDialog.Show()
	}

	applyFontSizeChange := func() {
		a.Settings().SetTheme(typographyTheme)
		if len(loadedData) == 0 {
			refreshViewport()
			return
		}
		if err := renderCurrentDocument(); err != nil {
			dialog.ShowError(err, w)
			return
		}
		a.Settings().SetTheme(typographyTheme)
	}

	increaseFontSize := func() {
		typographyTheme.IncreaseBodySize()
		applyFontSizeChange()
	}

	decreaseFontSize := func() {
		typographyTheme.DecreaseBodySize()
		applyFontSizeChange()
	}

	changeWordSpacing := func(next int) {
		if next < wordSpacingMin {
			next = wordSpacingMin
		}
		if next > wordSpacingMax {
			next = wordSpacingMax
		}
		wordSpacing = next
		if err := renderCurrentDocument(); err != nil {
			dialog.ShowError(err, w)
		}
	}

	showSettingsMenu := func() {
		menu := fyne.NewMenu("Settings",
			fyne.NewMenuItem("Load file...", openFile),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(fmt.Sprintf("Text size + (%.0f pt)", typographyTheme.BodySize()), increaseFontSize),
			fyne.NewMenuItem(fmt.Sprintf("Text size - (%.0f pt)", typographyTheme.BodySize()), decreaseFontSize),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(fmt.Sprintf("Word spacing + (x%d)", wordSpacing), func() {
				changeWordSpacing(wordSpacing + 1)
			}),
			fyne.NewMenuItem(fmt.Sprintf("Word spacing - (x%d)", wordSpacing), func() {
				changeWordSpacing(wordSpacing - 1)
			}),
		)

		popup := widget.NewPopUpMenu(menu, w.Canvas())
		popup.ShowAtRelativePosition(fyne.NewPos(0, controls.SettingsAnchor().Size().Height), controls.SettingsAnchor())
	}

	controls = NewControls(ControlActions{
		OnSettings: showSettingsMenu,
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
	}, engine.Speed())

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
		OnFontSizeUp:   increaseFontSize,
		OnFontSizeDown: decreaseFontSize,
	})

	w.SetContent(container.NewBorder(controls.View(), nil, nil, nil, scrollWithFade))
	w.ShowAndRun()
	return nil
}
