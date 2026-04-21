package main

import (
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/spyhere/re-peat/internal/configs"
	"github.com/spyhere/re-peat/internal/logging"
	"github.com/spyhere/re-peat/internal/state"
)

var (
	tag    = "dev"
	commit = "none"
	date   = "unknown"
)
var version = tag + "|" + commit + "|" + date

const logSize = 128 * 1024

func main() {
	lg := logging.NewLogger(version, logSize)
	defer func() {
		if r := recover(); r != nil {
			lg.Crash("r", r)
			os.Exit(1)
		}
	}()
	locale, err := configs.GetLocale()
	if err != nil {
		lg.Warn("Failed to get locale", "err", err)
	}
	window := new(app.Window)
	appState, err := state.NewAppState(window, locale, lg)
	if err != nil {
		lg.Crash("err", err)
		os.Exit(1)
	}
	repeatApp := newApp(&appState)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				lg.Crash("r", r)
				os.Exit(1)
			}
		}()
		window.Option(app.Title("re-peat"))
		window.Option(app.Size(unit.Dp(1000), unit.Dp(700)))
		err := run(window, repeatApp)
		if err != nil {
			lg.Warn("Window is prematurely closed", "err", err)
		}
		err = configs.SaveLocale(repeatApp.i18nSwitcher.Active.Lang.Tag())
		if err != nil {
			lg.Error("Failed to save i18n preference", err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window, repeatApp *App) error {
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			repeatApp.Layout(gtx, e)
			e.Frame(gtx.Ops)
		}
	}
}
