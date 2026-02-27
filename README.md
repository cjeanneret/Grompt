# grompt

[![CI](https://github.com/cjeanneret/Grompt/actions/workflows/pr-tests.yml/badge.svg?branch=main)](https://github.com/cjeanneret/Grompt/actions/workflows/pr-tests.yml)

![grompt logo](assets/icons/logo-512.png)

`grompt` is a desktop teleprompter app written in Go, using `Fyne` for the UI.
It loads Markdown or HTML files and displays them in a reading-friendly view with auto-scroll controls.

## Features

- Load `.md`, `.markdown`, `.html`, and `.htm` files
- Auto-scroll with adjustable speed
- Adjustable text size
- Adjustable word spacing
- Keyboard shortcuts for playback and typography controls
- Persistent user settings saved to a local config file

## Requirements

- Go `1.25.x` (module currently targets `go 1.25.6`)
- A desktop environment supported by `Fyne`
- CGO enabled on platforms where the GUI stack requires it

## Run

```bash
go run cmd/grompt/main.go
```

## Build

```bash
mkdir -p bin
podman run --rm \
  --security-opt label=disable \
  -v "$PWD:/workspace" \
  -w /workspace \
  golang:1.25-bookworm \
  bash -lc 'export PATH="/usr/local/go/bin:${PATH}" && apt-get update && apt-get install -y pkg-config libgl1-mesa-dev xorg-dev && go build -o bin/grompt cmd/grompt/main.go'
```

## App Icon Assets

The source logo is `assets/icons/logo.png`.
Generate transparent, content-cropped icon assets with:

```bash
go run ./cmd/logoassets
```

Generated files are written to `assets/icons`:

- `logo-clean.png` (transparent, cropped, square source for resizing)
- `logo-16.png`
- `logo-32.png`
- `logo-64.png`
- `logo-128.png`
- `logo-256.png`
- `logo-512.png`
- `logo-1024.png`

## Release Process

Releases are automated through GitHub Actions on tags matching `v*`.
When a new tag is pushed, GitHub builds and publishes binaries for Linux, macOS, and Windows.

```bash
git checkout main
git pull --ff-only origin main
git tag -a v0.0.1 -m "v0.0.1"
git push origin v0.0.1
```

## Usage

1. Open the app
2. Click the burger menu icon -> `Load file...`
3. Select a Markdown or HTML file
4. Use `Play` / `Pause` and speed controls
5. Adjust text size and word spacing from `Menu`
6. Use `Menu` -> `Exit` to close the app

## Keyboard Shortcuts

- `Space`: toggle play/pause
- `Arrow Up`: increase speed
- `Arrow Down`: decrease speed
- `+` or `=`: increase text size
- `-`: decrease text size

## Configuration File

`grompt` stores settings in:

```text
~/.config/grompt.conf
```

The format is simple `key=value` lines.
Blank lines and lines starting with `#` or `;` are treated as comments.

### Supported Options

- `speed` (float): auto-scroll speed in px/s
  - applied range: `20` to `300`
  - default: `20`
- `font_size` (float): main content font size in pt
  - applied range: `16` to `96`
  - default: `38`
- `word_spacing` (int): word spacing multiplier
  - applied range: `1` to `8`
  - default: `1`

### Example `grompt.conf`

```ini
speed=60
font_size=42
word_spacing=2
```

Invalid or out-of-range values are ignored or clamped, and the app can display a warning overlay at startup.

## Dependencies and Licenses

Major libraries currently used by this project:

| Library | Version | License |
| --- | --- | --- |
| `fyne.io/fyne/v2` | `v2.7.3` | BSD 3-Clause |
| `fyne.io/systray` | `v1.12.0` | Apache-2.0 |
| `github.com/go-gl/gl` | `v0.0.0-20231021071112-07e5d0ea2e71` | MIT |
| `github.com/go-gl/glfw/v3.3/glfw` | `v0.0.0-20240506104042-037f3cc74f2a` | BSD 3-Clause |
| `github.com/yuin/goldmark` | `v1.7.8` | MIT |
| `golang.org/x/net` | `v0.35.0` | BSD 3-Clause |
| `golang.org/x/text` | `v0.22.0` | BSD 3-Clause |
| `golang.org/x/image` | `v0.24.0` | BSD 3-Clause |
| `golang.org/x/sys` | `v0.30.0` | BSD 3-Clause |

Notes:

- This table focuses on core runtime dependencies.
- The full module list is available in `go.mod` / `go.sum`.
- Dependency licenses can change with version updates; always re-check before release.

## Project License

This project is licensed under the MIT License.
See `LICENSE` for the full text.

## AI-Assisted Development Disclosure

This project was produced with AI assistance.

- **IDE/Agent Platform:** Cursor IDE agent workflow
- **Primary AI model:** OpenAI GPT-5.3-Codex
- **Agent modes used:** `Agent` mode for implementation and iterative edits
- **Other modes:** `Plan` mode is available in Cursor for design/planning tasks; use depends on task complexity
- **Typical AI tools used during development:** code search, file reading/editing, terminal command execution, and patch-based file updates

Human review remains important for behavior validation, UX decisions, and release readiness.
