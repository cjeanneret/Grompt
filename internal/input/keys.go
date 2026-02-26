package input

import "fyne.io/fyne/v2"

type KeyActions struct {
	OnTogglePlayPause func()
	OnSpeedUp         func()
	OnSpeedDown       func()
	OnFontSizeUp      func()
	OnFontSizeDown    func()
}

func BindTeleprompterKeys(canvas fyne.Canvas, actions KeyActions) {
	canvas.SetOnTypedKey(func(event *fyne.KeyEvent) {
		switch event.Name {
		case fyne.KeySpace:
			if actions.OnTogglePlayPause != nil {
				actions.OnTogglePlayPause()
			}
		case fyne.KeyUp:
			if actions.OnSpeedUp != nil {
				actions.OnSpeedUp()
			}
		case fyne.KeyDown:
			if actions.OnSpeedDown != nil {
				actions.OnSpeedDown()
			}
		case fyne.KeyPlus, fyne.KeyEqual:
			if actions.OnFontSizeUp != nil {
				actions.OnFontSizeUp()
			}
		case fyne.KeyMinus:
			if actions.OnFontSizeDown != nil {
				actions.OnFontSizeDown()
			}
		}
	})
}
