package content

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

const (
	defaultWordSpacing = 1
	minWordSpacing     = 1
	maxWordSpacing     = 8
)

type RenderOptions struct {
	WordSpacing int
}

func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		WordSpacing: defaultWordSpacing,
	}
}

func NormalizeWordSpacing(value int) int {
	if value < minWordSpacing {
		return minWordSpacing
	}
	if value > maxWordSpacing {
		return maxWordSpacing
	}
	return value
}

func ApplyWordSpacing(object fyne.CanvasObject, spacing int) {
	richText, ok := object.(*widget.RichText)
	if !ok {
		return
	}

	normalized := NormalizeWordSpacing(spacing)
	for _, segment := range richText.Segments {
		applySegmentWordSpacing(segment, normalized)
	}
	richText.Refresh()
}

func applySegmentWordSpacing(segment widget.RichTextSegment, spacing int) {
	switch current := segment.(type) {
	case *widget.TextSegment:
		current.Text = stretchWords(current.Text, spacing)
	case *widget.HyperlinkSegment:
		current.Text = stretchWords(current.Text, spacing)
	case *widget.ParagraphSegment:
		for _, nested := range current.Segments() {
			applySegmentWordSpacing(nested, spacing)
		}
	case *widget.ListSegment:
		for _, nested := range current.Segments() {
			applySegmentWordSpacing(nested, spacing)
		}
	}
}

func stretchWords(value string, spacing int) string {
	if spacing <= 1 || value == "" {
		return value
	}
	return strings.ReplaceAll(value, " ", strings.Repeat(" ", spacing))
}
