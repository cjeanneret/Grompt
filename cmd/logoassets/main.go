package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"slices"

	xdraw "golang.org/x/image/draw"
)

var (
	defaultSizes = []int{16, 32, 64, 128, 256, 512, 1024}
)

func main() {
	inputPath := flag.String("input", "assets/icons/logo.png", "path to the source logo PNG")
	outputDir := flag.String("outdir", "assets/icons", "directory where generated icons are written")
	whiteThreshold := flag.Int("white-threshold", 26, "color distance threshold for white background removal")
	flag.Parse()

	sizes, err := parseSizes(defaultSizes)
	if err != nil {
		exitf("invalid icon sizes: %v", err)
	}

	source, err := os.Open(*inputPath)
	if err != nil {
		exitf("open input logo: %v", err)
	}
	defer source.Close()

	cfg, format, err := image.DecodeConfig(source)
	if err != nil {
		exitf("decode input config: %v", err)
	}
	if format != "png" {
		exitf("input must be PNG, got %q", format)
	}

	if _, err := source.Seek(0, 0); err != nil {
		exitf("rewind input logo: %v", err)
	}

	decoded, format, err := image.Decode(source)
	if err != nil {
		exitf("decode input image: %v", err)
	}
	if format != "png" {
		exitf("input must be PNG, got %q", format)
	}

	img := toNRGBA(decoded)
	originalHasAlpha := hasNonOpaquePixels(img)
	removed := removeWhiteBackground(img, uint8(*whiteThreshold))
	if removed == 0 {
		exitf("white background removal removed 0 pixels; adjust threshold or source")
	}

	bounds := nonTransparentBounds(img)
	if bounds.Empty() {
		exitf("content bounds empty after background removal")
	}
	cropped := cropToBounds(img, bounds)
	square := padToSquare(cropped)

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		exitf("create output directory: %v", err)
	}

	cleanPath := filepath.Join(*outputDir, "logo-clean.png")
	if err := writePNG(cleanPath, square); err != nil {
		exitf("write cleaned logo: %v", err)
	}

	for _, size := range sizes {
		resized := image.NewNRGBA(image.Rect(0, 0, size, size))
		xdraw.CatmullRom.Scale(resized, resized.Bounds(), square, square.Bounds(), draw.Over, nil)
		target := filepath.Join(*outputDir, fmt.Sprintf("logo-%d.png", size))
		if err := writePNG(target, resized); err != nil {
			exitf("write icon %d: %v", size, err)
		}
	}

	fmt.Printf("Source: %s (%dx%d)\n", *inputPath, cfg.Width, cfg.Height)
	fmt.Printf("Original alpha: %t\n", originalHasAlpha)
	fmt.Printf("Removed white bg pixels: %d\n", removed)
	fmt.Printf("Content bounds: %dx%d\n", cropped.Bounds().Dx(), cropped.Bounds().Dy())
	fmt.Printf("Square canvas: %dx%d\n", square.Bounds().Dx(), square.Bounds().Dy())
	fmt.Printf("Generated icons in %s: %v\n", *outputDir, sizes)
}

func parseSizes(defaults []int) ([]int, error) {
	if len(defaults) == 0 {
		return nil, errors.New("no sizes configured")
	}
	sizes := make([]int, 0, len(defaults))
	seen := map[int]struct{}{}
	for _, raw := range defaults {
		if raw <= 0 {
			return nil, fmt.Errorf("size %d must be > 0", raw)
		}
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		sizes = append(sizes, raw)
	}
	slices.Sort(sizes)
	return sizes, nil
}

func toNRGBA(src image.Image) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func hasNonOpaquePixels(img *image.NRGBA) bool {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if img.NRGBAAt(x, y).A < 0xff {
				return true
			}
		}
	}
	return false
}

func removeWhiteBackground(img *image.NRGBA, threshold uint8) int {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	if w == 0 || h == 0 {
		return 0
	}

	bg := averageCorners(img)
	visited := make([]bool, w*h)
	queue := make([]image.Point, 0, w*2+h*2)

	push := func(x, y int) {
		idx := y*w + x
		if visited[idx] {
			return
		}
		visited[idx] = true
		queue = append(queue, image.Pt(x, y))
	}

	for x := 0; x < w; x++ {
		push(x, 0)
		push(x, h-1)
	}
	for y := 1; y < h-1; y++ {
		push(0, y)
		push(w-1, y)
	}

	removed := 0
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		c := img.NRGBAAt(p.X, p.Y)
		if c.A == 0 || !isNear(bg, c, threshold) {
			continue
		}
		img.SetNRGBA(p.X, p.Y, color.NRGBA{R: c.R, G: c.G, B: c.B, A: 0})
		removed++

		if p.X > 0 {
			push(p.X-1, p.Y)
		}
		if p.X+1 < w {
			push(p.X+1, p.Y)
		}
		if p.Y > 0 {
			push(p.X, p.Y-1)
		}
		if p.Y+1 < h {
			push(p.X, p.Y+1)
		}
	}

	return removed
}

func averageCorners(img *image.NRGBA) color.NRGBA {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	points := []image.Point{
		image.Pt(0, 0),
		image.Pt(w-1, 0),
		image.Pt(0, h-1),
		image.Pt(w-1, h-1),
	}
	var r, g, b, a int
	for _, p := range points {
		c := img.NRGBAAt(p.X, p.Y)
		r += int(c.R)
		g += int(c.G)
		b += int(c.B)
		a += int(c.A)
	}
	return color.NRGBA{
		R: uint8(r / len(points)),
		G: uint8(g / len(points)),
		B: uint8(b / len(points)),
		A: uint8(a / len(points)),
	}
}

func isNear(a, b color.NRGBA, threshold uint8) bool {
	d := absInt(int(a.R)-int(b.R)) +
		absInt(int(a.G)-int(b.G)) +
		absInt(int(a.B)-int(b.B))
	return d <= int(threshold)*3
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func nonTransparentBounds(img *image.NRGBA) image.Rectangle {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	minX, minY := w, h
	maxX, maxY := -1, -1

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.NRGBAAt(x, y).A == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	if maxX < minX || maxY < minY {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func cropToBounds(src *image.NRGBA, bounds image.Rectangle) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Src)
	return dst
}

func padToSquare(src *image.NRGBA) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	size := w
	if h > size {
		size = h
	}
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	offsetX := (size - w) / 2
	offsetY := (size - h) / 2
	draw.Draw(dst, image.Rect(offsetX, offsetY, offsetX+w, offsetY+h), src, src.Bounds().Min, draw.Over)
	return dst
}

func writePNG(path string, img image.Image) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, img)
}

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
