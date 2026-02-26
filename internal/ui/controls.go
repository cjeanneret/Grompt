package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ControlActions struct {
	OnOpen      func()
	OnPlay      func()
	OnPause     func()
	OnSpeedUp   func()
	OnSpeedDown func()
}

type Controls struct {
	root       fyne.CanvasObject
	fileLabel  *widget.Label
	speedLabel *widget.Label
}

func NewControls(actions ControlActions, initialSpeed float64) *Controls {
	fileLabel := widget.NewLabel("No file loaded")
	speedLabel := widget.NewLabel(formatSpeed(initialSpeed))

	openButton := widget.NewButton("Open", actions.OnOpen)
	playButton := widget.NewButton("Play", actions.OnPlay)
	pauseButton := widget.NewButton("Pause", actions.OnPause)
	speedUpButton := widget.NewButton("Speed +", actions.OnSpeedUp)
	speedDownButton := widget.NewButton("Speed -", actions.OnSpeedDown)

	root := container.NewHBox(
		openButton,
		layout.NewSpacer(),
		fileLabel,
		layout.NewSpacer(),
		speedDownButton,
		speedLabel,
		speedUpButton,
		playButton,
		pauseButton,
	)

	return &Controls{
		root:       root,
		fileLabel:  fileLabel,
		speedLabel: speedLabel,
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

func formatSpeed(speed float64) string {
	return fmt.Sprintf("%.0f px/s", speed)
}
