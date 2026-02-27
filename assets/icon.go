package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icons/logo-512.png
var appIconPNG []byte

var appIconResource = fyne.NewStaticResource("logo-512.png", appIconPNG)

func AppIconResource() fyne.Resource {
	return appIconResource
}
