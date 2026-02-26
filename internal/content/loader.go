package content

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
)

var ErrUnsupportedFileType = errors.New("unsupported file type")

func LoadFromPath(path string) ([]byte, Format, error) {
	format, err := DetectFormat(path)
	if err != nil {
		return nil, "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read file: %w", err)
	}

	return data, format, nil
}

func DetectFormat(path string) (Format, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".md", ".markdown":
		return FormatMarkdown, nil
	case ".html", ".htm":
		return FormatHTML, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedFileType, ext)
	}
}
