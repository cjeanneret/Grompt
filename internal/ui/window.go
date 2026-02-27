package ui

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"grompt/assets"
	appconfig "grompt/internal/config"
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
	configPath, pathErr := appconfig.DefaultPath()
	configWarnings := make([]string, 0)
	if pathErr != nil {
		configWarnings = append(configWarnings, fmt.Sprintf("cannot resolve config path: %v", pathErr))
	}

	loadedSettings := appconfig.FileSettings{}
	if pathErr == nil {
		settings, warnings, loadErr := appconfig.Load(configPath)
		loadedSettings = settings
		configWarnings = append(configWarnings, warnings...)
		if loadErr != nil {
			configWarnings = append(configWarnings, fmt.Sprintf("cannot read config file: %v", loadErr))
		}
	}

	initialSpeed := scrollengine.DefaultSpeed
	if loadedSettings.Speed != nil {
		next := *loadedSettings.Speed
		if next < scrollengine.DefaultMinSpeed || next > scrollengine.DefaultMaxSpeed {
			configWarnings = append(configWarnings, fmt.Sprintf("speed %.0f out of range, clamped", next))
		}
		if next < scrollengine.DefaultMinSpeed {
			next = scrollengine.DefaultMinSpeed
		}
		if next > scrollengine.DefaultMaxSpeed {
			next = scrollengine.DefaultMaxSpeed
		}
		initialSpeed = next
	}

	initialFontSize := DefaultContentFontSize
	if loadedSettings.FontSize != nil {
		next := *loadedSettings.FontSize
		normalized := clampFontSize(next)
		if normalized != next {
			configWarnings = append(configWarnings, fmt.Sprintf("font_size %.0f out of range, clamped", next))
		}
		initialFontSize = normalized
	}

	initialWordSpacing := content.DefaultRenderOptions().WordSpacing
	if loadedSettings.WordSpacing != nil {
		next := *loadedSettings.WordSpacing
		normalized := content.NormalizeWordSpacing(next)
		if normalized != next {
			configWarnings = append(configWarnings, fmt.Sprintf("word_spacing %d out of range, clamped", next))
		}
		initialWordSpacing = normalized
	}

	a := app.NewWithID("com.grompt.app")
	a.SetIcon(assets.AppIconResource())
	typographyTheme := NewTypographyTheme(initialFontSize)
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
	engine.SetSpeed(initialSpeed)

	var settingsWriter *appconfig.AsyncWriter
	if pathErr == nil {
		settingsWriter = appconfig.NewAsyncWriter(configPath)
		defer settingsWriter.Close()
	}

	var controls *Controls
	wordSpacing := initialWordSpacing
	var loadedData []byte
	var loadedFormat content.Format
	var loadedFileName string

	saveSettings := func() {
		if settingsWriter == nil {
			return
		}
		settingsWriter.Save(appconfig.Settings{
			Speed:       engine.Speed(),
			FontSize:    typographyTheme.BodySize(),
			WordSpacing: wordSpacing,
		})
	}

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
		saveSettings()
	}

	decreaseFontSize := func() {
		typographyTheme.DecreaseBodySize()
		applyFontSizeChange()
		saveSettings()
	}

	changeWordSpacing := func(next int) {
		wordSpacing = content.NormalizeWordSpacing(next)
		if err := renderCurrentDocument(); err != nil {
			dialog.ShowError(err, w)
			return
		}
		saveSettings()
	}

	showSettingsMenu := func() {
		menu := fyne.NewMenu("Menu",
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
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Exit", func() {
				a.Quit()
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
			saveSettings()
		},
		OnSpeedDown: func() {
			controls.SetSpeed(engine.SpeedDown())
			saveSettings()
		},
	}, engine.Speed())

	input.BindTeleprompterKeys(w.Canvas(), input.KeyActions{
		OnTogglePlayPause: func() {
			engine.Toggle()
		},
		OnSpeedUp: func() {
			controls.SetSpeed(engine.SpeedUp())
			saveSettings()
		},
		OnSpeedDown: func() {
			controls.SetSpeed(engine.SpeedDown())
			saveSettings()
		},
		OnFontSizeUp:   increaseFontSize,
		OnFontSizeDown: decreaseFontSize,
	})

	w.SetContent(container.NewBorder(controls.View(), nil, nil, nil, scrollWithFade))
	if len(configWarnings) > 0 {
		showConfigWarningOverlay(w, configWarnings)
	}
	w.ShowAndRun()
	return nil
}

func showConfigWarningOverlay(w fyne.Window, warnings []string) {
	visibleWarnings := warnings
	if len(visibleWarnings) > 4 {
		visibleWarnings = append(visibleWarnings[:4], fmt.Sprintf("... and %d more", len(warnings)-4))
	}

	message := widget.NewLabel(
		"Some config values were ignored:\n- " + strings.Join(visibleWarnings, "\n- "),
	)
	message.Wrapping = fyne.TextWrapWord

	var popup *widget.PopUp
	closeButton := widget.NewButton("×", func() {
		if popup != nil {
			popup.Hide()
		}
	})
	closeButton.Importance = widget.LowImportance

	header := container.NewBorder(nil, nil, nil, closeButton, widget.NewLabel("Configuration warning"))
	contentView := container.NewBorder(header, nil, nil, nil, message)

	popup = widget.NewPopUp(contentView, w.Canvas())
	popup.Resize(fyne.NewSize(460, 160))
	popup.Move(fyne.NewPos(16, 16))
	popup.Show()
}
