package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ControlActions struct {
	OnOpen         func()
	OnPlay         func()
	OnPause        func()
	OnSpeedUp      func()
	OnSpeedDown    func()
	OnFontSizeUp   func()
	OnFontSizeDown func()
}

type Controls struct {
	root          fyne.CanvasObject
	fileLabel     *widget.Label
	speedLabel    *widget.Label
	fontSizeLabel *widget.Label
}

func NewControls(actions ControlActions, initialSpeed float64, initialFontSize float32) *Controls {
	fileLabel := widget.NewLabel("No file loaded")
	speedLabel := widget.NewLabel(formatSpeed(initialSpeed))
	fontSizeLabel := widget.NewLabel(formatFontSize(initialFontSize))

	openButton := widget.NewButton("Open", actions.OnOpen)
	playButton := widget.NewButton("Play", actions.OnPlay)
	pauseButton := widget.NewButton("Pause", actions.OnPause)
	speedUpButton := widget.NewButton("Speed +", actions.OnSpeedUp)
	speedDownButton := widget.NewButton("Speed -", actions.OnSpeedDown)
	fontSizeUpButton := widget.NewButton("Font +", actions.OnFontSizeUp)
	fontSizeDownButton := widget.NewButton("Font -", actions.OnFontSizeDown)

	root := container.NewHBox(
		openButton,
		layout.NewSpacer(),
		fileLabel,
		layout.NewSpacer(),
		fontSizeDownButton,
		fontSizeLabel,
		fontSizeUpButton,
		speedDownButton,
		speedLabel,
		speedUpButton,
		playButton,
		pauseButton,
	)

	return &Controls{
		root:          root,
		fileLabel:     fileLabel,
		speedLabel:    speedLabel,
		fontSizeLabel: fontSizeLabel,
	}
}

func (c *Controls) View() fyne.CanvasObject {
	return c.root
}

func (c *Controls) SetFileName(name string) {
	c.fileLabel.SetText(name)
}

func (c *Controls) SetSpeed(speed float64) {
	c.speedLabel.SetText(formatSpeed(speed))
}

func (c *Controls) SetFontSize(size float32) {
	c.fontSizeLabel.SetText(formatFontSize(size))
}

func formatSpeed(speed float64) string {
	return fmt.Sprintf("%.0f px/s", speed)
}

func formatFontSize(size float32) string {
	return fmt.Sprintf("%.0f pt", size)
}
