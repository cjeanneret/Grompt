package content

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	ThemeSizeContentBody       fyne.ThemeSizeName = "grompt.content.body"
	ThemeSizeContentHeading    fyne.ThemeSizeName = "grompt.content.heading"
	ThemeSizeContentSubheading fyne.ThemeSizeName = "grompt.content.subheading"
)

func ApplyTypography(object fyne.CanvasObject) {
	richText, ok := object.(*widget.RichText)
	if !ok {
		return
	}

	for _, segment := range richText.Segments {
		applySegmentTypography(segment)
	}
	richText.Refresh()
}

func applySegmentTypography(segment widget.RichTextSegment) {
	switch current := segment.(type) {
	case *widget.TextSegment:
		switch current.Style.SizeName {
		case theme.SizeNameHeadingText:
			current.Style.SizeName = ThemeSizeContentHeading
		case theme.SizeNameSubHeadingText:
			current.Style.SizeName = ThemeSizeContentSubheading
		default:
			current.Style.SizeName = ThemeSizeContentBody
		}
	case *widget.ParagraphSegment:
		for _, nested := range current.Segments() {
			applySegmentTypography(nested)
		}
	case *widget.ListSegment:
		for _, nested := range current.Segments() {
			applySegmentTypography(nested)
		}
	}
}
