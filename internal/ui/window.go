package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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

	content := widget.NewRichTextFromMarkdown(initialMessage)
	content.Wrapping = fyne.TextWrapWord

	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(640, 400))

	openButton := widget.NewButton("Open", func() {
		fmt.Println("Open action will be implemented in next step.")
	})
	playButton := widget.NewButton("Play", nil)
	pauseButton := widget.NewButton("Pause", nil)
	speedUpButton := widget.NewButton("Speed +", nil)
	speedDownButton := widget.NewButton("Speed -", nil)

	controls := container.NewHBox(
		openButton,
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
