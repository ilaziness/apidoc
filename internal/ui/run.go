// Package ui is the go-gui presentation layer.
//
// Files follow the window regions, not feature layers:
//
//	run.go      window bootstrap
//	app.go      session state (selection, drafts, dirty)
//	layout.go   icon toolbar + page body
//	home.go     collection tiles
//	sidebar.go  current collection tree
//	editor.go   node workspace (name/save, folder, markdown)
//	http.go     HTTP preview / edit / debug
//	panel.go    settings, import, trash, environment (dialogs or trash page)
//	colors.go   UI color constants (dialog chrome, etc.)
//	text.go     Key:Value codecs used by editors
//
// Views call service; they do not open SQLite or send HTTP themselves.
package ui

import (
	"context"

	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/backend"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/service"
)

// Run starts the desktop window.
func Run(cfg *config.Config, svc *service.Service) error {
	applyWindowTheme(cfg)
	ctx := context.Background()
	_ = svc.PurgeTrash(ctx)

	app := newApp(svc)
	app.refreshCollections(ctx)
	app.loadSettingsDraft(cfg)

	w := gui.NewWindow(gui.WindowCfg{
		State:  app,
		Title:  app.T().AppTitle,
		Width:  1280,
		Height: 800,
		OnInit: func(w *gui.Window) { w.SetView(mainView) },
	})
	backend.Run(w)
	return nil
}

func applyWindowTheme(cfg *config.Config) {
	switch cfg.Theme {
	case config.ThemeDark:
		gui.SetTheme(gui.ThemeDark)
	case config.ThemeLight:
		gui.SetTheme(gui.ThemeLight)
	default:
		gui.SetTheme(gui.ThemeLight)
	}
}

func appOf(w *gui.Window) *App { return gui.State[App](w) }

func appCtx(ctx gui.EventCtx) *App { return gui.State[App](ctx.Window) }
