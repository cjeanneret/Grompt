package ui

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"grompt/internal/content"
)

const (
	DefaultContentFontSize float32 = 30
	MinContentFontSize     float32 = 16
	MaxContentFontSize     float32 = 96
	ContentFontStep        float32 = 2
)

type TypographyTheme struct {
	mu       sync.RWMutex
	base     fyne.Theme
	bodySize float32
}

func NewTypographyTheme(bodySize float32) *TypographyTheme {
	size := clampFontSize(bodySize)
	return &TypographyTheme{
		base:     theme.DefaultTheme(),
		bodySize: size,
	}
}

func (t *TypographyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *TypographyTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *TypographyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *TypographyTheme) Size(name fyne.ThemeSizeName) float32 {
	t.mu.RLock()
	body := t.bodySize
	t.mu.RUnlock()

	switch name {
	case content.ThemeSizeContentBody:
		return body
	case content.ThemeSizeContentSubheading:
		return body * 1.1
	case content.ThemeSizeContentHeading:
		return body * 1.2
	default:
		return t.base.Size(name)
	}
}

func (t *TypographyTheme) BodySize() float32 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.bodySize
}

func (t *TypographyTheme) IncreaseBodySize() float32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bodySize = clampFontSize(t.bodySize + ContentFontStep)
	return t.bodySize
}

func (t *TypographyTheme) DecreaseBodySize() float32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bodySize = clampFontSize(t.bodySize - ContentFontStep)
	return t.bodySize
}

func estimatedLineHeight(fontSize float32) float32 {
	return fontSize * 1.35
}

func clampFontSize(size float32) float32 {
	if size < MinContentFontSize {
		return MinContentFontSize
	}
	if size > MaxContentFontSize {
		return MaxContentFontSize
	}
	return size
}
