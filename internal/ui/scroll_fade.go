package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

const (
	clearReadingLines = float32(3)
	chevronLineHeight = float32(3)
	minSideGutter     = float32(40)
)

type scrollFadeLayout struct {
	clearLines         float32
	lineHeightProvider func() float32
}

func NewScrollWithFade(scroll *container.Scroll, lineHeightProvider func() float32) fyne.CanvasObject {
	background := color.NRGBAModel.Convert(theme.Color(theme.ColorNameBackground)).(color.NRGBA)
	topGradient := canvas.NewVerticalGradient(withAlpha(background, 230), withAlpha(background, 0))
	bottomGradient := canvas.NewVerticalGradient(withAlpha(background, 0), withAlpha(background, 230))
	leftChevron := canvas.NewText(">", withThemeAlpha(theme.Color(theme.ColorNameForeground), 220))
	rightChevron := canvas.NewText("<", withThemeAlpha(theme.Color(theme.ColorNameForeground), 220))

	return container.New(&scrollFadeLayout{
		clearLines:         clearReadingLines,
		lineHeightProvider: lineHeightProvider,
	}, scroll, topGradient, bottomGradient, leftChevron, rightChevron)
}

func (l *scrollFadeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	scroll := objects[0]

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

	chevronSize := lineHeight * chevronLineHeight
	if chevronSize < 16 {
		chevronSize = 16
	}

	sideGutter := chevronSize*0.95 + minSideGutter
	maxGutter := (size.Width - 120) / 2
	if maxGutter < minSideGutter {
		maxGutter = minSideGutter
	}
	if sideGutter > maxGutter {
		sideGutter = maxGutter
	}
	if sideGutter < minSideGutter {
		sideGutter = minSideGutter
	}

	scrollX := sideGutter
	scrollWidth := size.Width - (2 * sideGutter)
	if scrollWidth < 120 {
		scrollWidth = 120
		scrollX = (size.Width - scrollWidth) / 2
		if scrollX < 0 {
			scrollX = 0
			scrollWidth = size.Width
		}
	}

	scroll.Move(fyne.NewPos(scrollX, 0))
	scroll.Resize(fyne.NewSize(scrollWidth, size.Height))

	if len(objects) < 3 {
		return
	}

	topGradient := objects[1]
	bottomGradient := objects[2]

	topGradient.Move(fyne.NewPos(scrollX, 0))
	topGradient.Resize(fyne.NewSize(scrollWidth, fadeHeight))

	bottomGradient.Move(fyne.NewPos(scrollX, size.Height-fadeHeight))
	bottomGradient.Resize(fyne.NewSize(scrollWidth, fadeHeight))

	if len(objects) < 5 {
		return
	}

	centerY := fadeHeight + (clearBandHeight / 2)

	leftChevron, ok := objects[3].(*canvas.Text)
	if ok {
		leftChevron.TextSize = chevronSize
		leftChevron.Refresh()
		leftSize := leftChevron.MinSize()
		leftChevron.Move(fyne.NewPos(scrollX-leftSize.Width-20, centerY-(leftSize.Height/2)))
	}

	rightChevron, ok := objects[4].(*canvas.Text)
	if ok {
		rightChevron.TextSize = chevronSize
		rightChevron.Refresh()
		rightSize := rightChevron.MinSize()
		rightChevron.Move(fyne.NewPos(scrollX+scrollWidth+20, centerY-(rightSize.Height/2)))
	}
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

func withThemeAlpha(c color.Color, alpha uint8) color.NRGBA {
	converted := color.NRGBAModel.Convert(c).(color.NRGBA)
	converted.A = alpha
	return converted
}
