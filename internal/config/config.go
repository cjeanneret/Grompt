package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultFileName  = "grompt.conf"
	defaultWriteWait = 250 * time.Millisecond
)

type FileSettings struct {
	Speed       *float64
	FontSize    *float32
	WordSpacing *int
}

type Settings struct {
	Speed       float64
	FontSize    float32
	WordSpacing int
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", defaultFileName), nil
}

func Load(path string) (FileSettings, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return FileSettings{}, nil, nil
		}
		return FileSettings{}, nil, err
	}
	defer file.Close()

	settings := FileSettings{}
	var warnings []string

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			warnings = append(warnings, fmt.Sprintf("line %d ignored (expected key=value)", lineNo))
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])
		if value == "" {
			warnings = append(warnings, fmt.Sprintf("%s is empty and was ignored", key))
			continue
		}

		switch key {
		case "speed":
			parsed, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil {
				warnings = append(warnings, fmt.Sprintf("invalid speed=%q ignored", value))
				continue
			}
			settings.Speed = &parsed
		case "font_size":
			parsed, parseErr := strconv.ParseFloat(value, 32)
			if parseErr != nil {
				warnings = append(warnings, fmt.Sprintf("invalid font_size=%q ignored", value))
				continue
			}
			asFloat32 := float32(parsed)
			settings.FontSize = &asFloat32
		case "word_spacing":
			parsed, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				warnings = append(warnings, fmt.Sprintf("invalid word_spacing=%q ignored", value))
				continue
			}
			settings.WordSpacing = &parsed
		default:
			warnings = append(warnings, fmt.Sprintf("unknown setting %q ignored", key))
		}
	}

	if err := scanner.Err(); err != nil {
		return FileSettings{}, warnings, err
	}

	return settings, warnings, nil
}

type AsyncWriter struct {
	path    string
	wait    time.Duration
	updates chan Settings
	stopCh  chan struct{}
	doneCh  chan struct{}
	once    sync.Once
}

func NewAsyncWriter(path string) *AsyncWriter {
	writer := &AsyncWriter{
		path:    path,
		wait:    defaultWriteWait,
		updates: make(chan Settings, 1),
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}

	go writer.loop()
	return writer
}

func (w *AsyncWriter) Save(settings Settings) {
	select {
	case w.updates <- settings:
	default:
		select {
		case <-w.updates:
		default:
		}
		select {
		case w.updates <- settings:
		default:
		}
	}
}

func (w *AsyncWriter) Close() {
	w.once.Do(func() {
		close(w.stopCh)
		<-w.doneCh
	})
}

func (w *AsyncWriter) loop() {
	defer close(w.doneCh)

	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}

	var pending *Settings

	for {
		select {
		case update := <-w.updates:
			next := update
			pending = &next
			resetTimer(timer, w.wait)
		case <-timer.C:
			if pending != nil {
				_ = writeAtomic(w.path, *pending)
				pending = nil
			}
		case <-w.stopCh:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			if pending != nil {
				_ = writeAtomic(w.path, *pending)
			}
			return
		}
	}
}

func resetTimer(timer *time.Timer, delay time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(delay)
}

func writeAtomic(path string, settings Settings) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "grompt-*.tmp")
	if err != nil {
		return err
	}

	content := fmt.Sprintf("speed=%.0f\nfont_size=%.0f\nword_spacing=%d\n", settings.Speed, settings.FontSize, settings.WordSpacing)
	if _, err = tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}
