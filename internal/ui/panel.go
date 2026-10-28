package ui

import (
	"context"
	"fmt"

	"github.com/go-gui-org/go-gui/gui"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/openapi"
	"github.com/ilaziness/apidoc/internal/service"
)

func importSource(a *App) (openapi.Source, error) {
	switch {
	case a.ImportPath != "" && a.ImportURL == "":
		return openapi.Source{FilePath: a.ImportPath}, nil
	case a.ImportURL != "" && a.ImportPath == "":
		return openapi.Source{URL: a.ImportURL}, nil
	default:
		return openapi.Source{}, fmt.Errorf("set exactly one of file path or URL")
	}
}

func importView(app *App) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Padding: gui.PaddingSmall,
		Content: []gui.View{
			gui.Label(cat.ImportOpenAPI, gui.CurrentTheme().TextStyleTitle),
			gui.Row(gui.ContainerCfg{
				Spacing: gui.SpacingSmall,
				VAlign:  gui.VAlignMiddle,
				Sizing:  gui.FillFit,
				Content: []gui.View{
					gui.Input(gui.InputCfg{
						ID:     "import-path",
						Label:  cat.FilePath,
						Text:   app.ImportPath,
						Sizing: gui.FillFit,
						OnTextChanged: func(s string, ctx gui.EventCtx) {
							appCtx(ctx).ImportPath = s
						},
					}),
					gui.TextButton(cat.FilePath, func(ctx gui.EventCtx) {
						openImport(ctx.Window)
					}),
				},
			}),
			gui.Input(gui.InputCfg{
				ID:    "import-url",
				Label: "URL",
				Text:  app.ImportURL,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					appCtx(ctx).ImportURL = s
				},
			}),
			gui.TextButton(cat.ImportNew, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				src, err := importSource(a)
				if err != nil {
					a.fail(err)
					return
				}
				col, err := a.Svc.ImportOpenAPI(context.Background(), service.ImportOpenAPIOptions{Source: src})
				if err != nil {
					a.fail(err)
					return
				}
				a.refreshCollections(context.Background())
				a.selectCollection(context.Background(), col.ID)
				a.Status = cat.Saved
				toastOK(ctx.Window, cat.Saved)
			}),
			gui.TextButton(cat.ImportCurrent, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				if a.SelectedCollectionID == 0 {
					a.Status = a.T().NoSelection
					return
				}
				src, err := importSource(a)
				if err != nil {
					a.fail(err)
					return
				}
				_, err = a.Svc.ImportOpenAPI(context.Background(), service.ImportOpenAPIOptions{
					Source:           src,
					IntoCollectionID: a.SelectedCollectionID,
				})
				if err != nil {
					a.fail(err)
					return
				}
				a.refreshTree(context.Background())
				a.Panel = panelEditor
				a.Status = cat.Saved
				toastOK(ctx.Window, cat.Saved)
			}),
			gui.TextButton(cat.Cancel, closeEditor),
		},
	})
}

func settingsView(app *App) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Padding: gui.PaddingSmall,
		Content: []gui.View{
			gui.Label(cat.Settings, gui.CurrentTheme().TextStyleTitle),
			gui.Select(gui.SelectCfg{
				ID:    "set-lang",
				Label: cat.Language,
				Options: []gui.SelectOption{
					gui.NewSelectOption("English", config.LangEN),
					gui.NewSelectOption("简体中文", config.LangZHHans),
					gui.NewSelectOption("繁體中文", config.LangZHHant),
				},
				Selected: []string{app.SetLanguage},
				OnSelect: func(vals []string, ctx gui.EventCtx) {
					if len(vals) > 0 {
						appCtx(ctx).SetLanguage = vals[0]
					}
				},
			}),
			gui.Select(gui.SelectCfg{
				ID:    "set-theme",
				Label: cat.Theme,
				Options: []gui.SelectOption{
					gui.NewSelectOption(cat.ThemeLight, config.ThemeLight),
					gui.NewSelectOption(cat.ThemeDark, config.ThemeDark),
					gui.NewSelectOption(cat.ThemeSystem, config.ThemeSystem),
				},
				Selected: []string{app.SetTheme},
				OnSelect: func(vals []string, ctx gui.EventCtx) {
					if len(vals) > 0 {
						appCtx(ctx).SetTheme = vals[0]
					}
				},
			}),
			gui.Select(gui.SelectCfg{
				ID:    "set-proxy",
				Label: cat.Proxy,
				Options: []gui.SelectOption{
					gui.NewSelectOption(cat.ProxySystem, config.ProxySystem),
					gui.NewSelectOption(cat.ProxyOff, config.ProxyOff),
					gui.NewSelectOption(cat.ProxyManual, config.ProxyManual),
				},
				Selected: []string{app.SetProxyMode},
				OnSelect: func(vals []string, ctx gui.EventCtx) {
					if len(vals) > 0 {
						appCtx(ctx).SetProxyMode = vals[0]
					}
				},
			}),
			gui.Input(gui.InputCfg{
				ID:    "proxy-url",
				Label: cat.ProxyManual,
				Text:  app.SetProxyURL,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					appCtx(ctx).SetProxyURL = s
				},
			}),
			gui.Input(gui.InputCfg{
				ID:    "data-dir",
				Label: cat.DataDir,
				Text:  app.SetDataDir,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					appCtx(ctx).SetDataDir = s
				},
			}),
			gui.TextButton(cat.Save, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				cfg := a.Svc.Config()
				cfg.Language = a.SetLanguage
				cfg.Theme = a.SetTheme
				cfg.Proxy.Mode = a.SetProxyMode
				cfg.Proxy.Manual = a.SetProxyURL
				if a.SetDataDir != "" && a.SetDataDir != cfg.DataDir {
					if err := config.CopyDataDir(cfg.DataDir, a.SetDataDir); err != nil {
						a.fail(err)
						return
					}
					cfg.DataDir = a.SetDataDir
					if err := a.Svc.SaveConfig(); err != nil {
						a.fail(err)
						return
					}
					a.Status = "Data directory copied; please restart the app"
					a.Panel = panelEditor
					return
				}
				if err := a.Svc.SaveConfig(); err != nil {
					a.fail(err)
					return
				}
				applyWindowTheme(cfg)
				a.Panel = panelEditor
				a.Status = cat.Saved
			}),
			gui.TextButton(cat.Cancel, closeEditor),
		},
	})
}

func trashView(app *App) gui.View {
	cat := app.T()
	cols, _ := app.Svc.ListTrashCollections(context.Background())
	nodes, _ := app.Svc.ListTrashNodes(context.Background())
	items := []gui.View{gui.Label(cat.Trash, gui.CurrentTheme().TextStyleTitle)}
	for _, c := range cols {
		id := c.ID
		name := c.Name
		items = append(items, gui.Row(gui.ContainerCfg{
			Spacing: gui.SpacingSmall,
			Content: []gui.View{
				gui.Label(cat.Collections+": "+name, gui.TextStyle{}),
				gui.TextButton(cat.Restore, func(ctx gui.EventCtx) {
					a := appCtx(ctx)
					if err := a.Svc.RestoreCollection(context.Background(), id); err != nil {
						a.fail(err)
						return
					}
					a.refreshCollections(context.Background())
				}),
			},
		}))
	}
	for _, n := range nodes {
		id := n.ID
		name := n.Name
		kind := n.Kind
		items = append(items, gui.Row(gui.ContainerCfg{
			Spacing: gui.SpacingSmall,
			Content: []gui.View{
				gui.Label(kind+": "+name, gui.TextStyle{}),
				gui.TextButton(cat.Restore, func(ctx gui.EventCtx) {
					a := appCtx(ctx)
					if err := a.Svc.RestoreNode(context.Background(), id); err != nil {
						a.fail(err)
						return
					}
					a.refreshTree(context.Background())
				}),
			},
		}))
	}
	items = append(items,
		gui.TextButtonVariant(cat.EmptyTrash, gui.ButtonDanger, func(ctx gui.EventCtx) {
			a := appCtx(ctx)
			if err := a.Svc.EmptyTrash(context.Background()); err != nil {
				a.fail(err)
				return
			}
			a.Status = cat.Saved
		}),
		gui.TextButton(cat.Cancel, closeEditor),
	)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Padding: gui.PaddingSmall,
		Content: items,
	})
}

func envView(app *App) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Padding: gui.PaddingSmall,
		Content: []gui.View{
			gui.Label(cat.EditEnv, gui.CurrentTheme().TextStyleTitle),
			gui.Input(gui.InputCfg{
				ID:    "env-name",
				Label: cat.Rename,
				Text:  app.DraftEnvName,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					appCtx(ctx).DraftEnvName = s
				},
			}),
			gui.Input(gui.InputCfg{
				ID:     "env-vars",
				Label:  cat.Variables + "  (Key:Value)",
				Mode:   gui.InputMultiline,
				Text:   app.DraftEnvVars,
				Sizing: gui.FillFill,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					appCtx(ctx).DraftEnvVars = s
				},
			}),
			gui.TextButton(cat.Save, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				if a.SelectedEnvID == 0 {
					return
				}
				e, err := a.Svc.GetEnvironment(context.Background(), a.SelectedEnvID)
				if err != nil {
					a.fail(err)
					return
				}
				e.Name = a.DraftEnvName
				if err := a.Svc.UpdateEnvironment(context.Background(), e, decodeMap(a.DraftEnvVars)); err != nil {
					a.fail(err)
					return
				}
				a.refreshTree(context.Background())
				a.Panel = panelEditor
				a.Status = cat.Saved
			}),
			gui.TextButtonVariant(cat.Delete, gui.ButtonDanger, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				if a.SelectedEnvID == 0 {
					return
				}
				if err := a.Svc.DeleteEnvironment(context.Background(), a.SelectedEnvID); err != nil {
					a.fail(err)
					return
				}
				a.SelectedEnvID = 0
				a.refreshTree(context.Background())
				if len(a.Envs) > 0 {
					a.SelectedEnvID = a.Envs[0].ID
				}
				a.syncEnvDraft()
				a.Panel = panelEditor
			}),
			gui.TextButton(cat.Cancel, closeEditor),
		},
	})
}
