package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

const clearReadingLines = float32(2.5)

type scrollFadeLayout struct {
	clearLines         float32
	lineHeightProvider func() float32
}

func NewScrollWithFade(scroll *container.Scroll, lineHeightProvider func() float32) fyne.CanvasObject {
	background := color.NRGBAModel.Convert(theme.Color(theme.ColorNameBackground)).(color.NRGBA)
	topGradient := canvas.NewVerticalGradient(withAlpha(background, 230), withAlpha(background, 0))
	bottomGradient := canvas.NewVerticalGradient(withAlpha(background, 0), withAlpha(background, 230))

	return container.New(&scrollFadeLayout{
		clearLines:         clearReadingLines,
		lineHeightProvider: lineHeightProvider,
	}, scroll, topGradient, bottomGradient)
}

func (l *scrollFadeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	scroll := objects[0]
	scroll.Move(fyne.Position{})
	scroll.Resize(size)

	if len(objects) < 3 {
		return
	}

	lineHeight := float32(1)
	if l.lineHeightProvider != nil {
		if provided := l.lineHeightProvider(); provided > 0 {
			lineHeight = provided
		}
	}

	clearBandHeight := lineHeight * l.clearLines
	if clearBandHeight > size.Height {
		clearBandHeight = size.Height
	}
	fadeHeight := (size.Height - clearBandHeight) / 2
	if fadeHeight < 0 {
		fadeHeight = 0
	}

	topGradient := objects[1]
	bottomGradient := objects[2]

	topGradient.Move(fyne.NewPos(0, 0))
	topGradient.Resize(fyne.NewSize(size.Width, fadeHeight))

	bottomGradient.Move(fyne.NewPos(0, size.Height-fadeHeight))
	bottomGradient.Resize(fyne.NewSize(size.Width, fadeHeight))
}

func (l *scrollFadeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	return objects[0].MinSize()
}

func withAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = alpha
	return c
}
