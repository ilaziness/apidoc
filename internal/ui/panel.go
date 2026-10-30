package ui

import (
	"context"
	"fmt"
	"strings"

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

func openImport(w *gui.Window) {
	askUnsaved(w, func() {
		a := appOf(w)
		a.ImportPath = ""
		a.ImportURL = ""
		customDialog(w, "import-dialog", 520, 0, func(win *gui.Window) gui.View {
			return importView(appOf(win))
		})
	})
}

func openNewProjectDialog(w *gui.Window) {
	askUnsaved(w, func() {
		appOf(w).NewProjectName = ""
		customDialog(w, "new-project-dialog", 400, 0, func(win *gui.Window) gui.View {
			return newProjectView(appOf(win))
		})
	})
}

func newProjectView(app *App) gui.View {
	cat := app.T()
	return dialogShell("new-project", cat.NewCollection, []gui.View{
		gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Padding: gui.PaddingMedium,
			Spacing: gui.SpacingSmall,
			Content: []gui.View{
				gui.Input(gui.InputCfg{
					ID:    "new-project-name",
					Label: cat.Name,
					Text:  app.NewProjectName,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						appCtx(ctx).NewProjectName = s
					},
				}),
				gui.Row(gui.ContainerCfg{
					Sizing: gui.FillFit,
					HAlign: gui.HAlignEnd,
					Content: []gui.View{
						gui.TextButtonVariant(cat.OK, gui.ButtonPrimary, func(ctx gui.EventCtx) {
							createProject(ctx.Window)
						}),
					},
				}),
			},
		}),
	})
}

func createProject(w *gui.Window) {
	a := appOf(w)
	name := strings.TrimSpace(a.NewProjectName)
	if name == "" {
		return
	}
	c, err := a.Svc.CreateCollection(context.Background(), name)
	if err != nil {
		toastErr(w, err.Error())
		return
	}
	w.DialogDismiss()
	a.refreshCollections(context.Background())
	a.selectCollection(context.Background(), c.ID)
}

func pickImportFile(w *gui.Window) {
	cat := appOf(w).T()
	w.NativeOpenDialog(gui.NativeOpenDialogCfg{
		Title: cat.ImportOpenAPI,
		Filters: []gui.NativeFileFilter{
			{Name: "OpenAPI", Extensions: []string{"json", "yaml", "yml"}},
		},
		OnDone: func(r gui.NativeDialogResult, win *gui.Window) {
			if r.Status != gui.DialogOK {
				return
			}
			paths := r.PathStrings()
			if len(paths) == 0 {
				return
			}
			a := appOf(win)
			a.ImportPath = paths[0]
			a.ImportURL = ""
		},
	})
}

func importView(app *App) gui.View {
	cat := app.T()
	return dialogShell("import", cat.ImportOpenAPI, []gui.View{
		gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Spacing: gui.SpacingSmall,
			Padding: gui.PaddingMedium,
			Content: []gui.View{
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
								a := appCtx(ctx)
								a.ImportPath = s
								if s != "" {
									a.ImportURL = ""
								}
							},
						}),
						gui.TextButton(cat.Browse, func(ctx gui.EventCtx) {
							pickImportFile(ctx.Window)
						}),
					},
				}),
				gui.Input(gui.InputCfg{
					ID:    "import-url",
					Label: cat.URL,
					Text:  app.ImportURL,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						a.ImportURL = s
						if s != "" {
							a.ImportPath = ""
						}
					},
				}),
				gui.TextButton(cat.ImportNew, func(ctx gui.EventCtx) {
					a := appCtx(ctx)
					src, err := importSource(a)
					if err != nil {
						toastErr(ctx.Window, err.Error())
						return
					}
					col, err := a.Svc.ImportOpenAPI(context.Background(), service.ImportOpenAPIOptions{Source: src})
					if err != nil {
						toastErr(ctx.Window, err.Error())
						return
					}
					a.refreshCollections(context.Background())
					a.selectCollection(context.Background(), col.ID)
					ctx.Window.DialogDismiss()
					toastOK(ctx.Window, cat.Saved)
				}),
				gui.TextButton(cat.ImportCurrent, func(ctx gui.EventCtx) {
					a := appCtx(ctx)
					if a.SelectedCollectionID == 0 {
						toastErr(ctx.Window, a.T().NoSelection)
						return
					}
					src, err := importSource(a)
					if err != nil {
						toastErr(ctx.Window, err.Error())
						return
					}
					_, err = a.Svc.ImportOpenAPI(context.Background(), service.ImportOpenAPIOptions{
						Source:           src,
						IntoCollectionID: a.SelectedCollectionID,
					})
					if err != nil {
						toastErr(ctx.Window, err.Error())
						return
					}
					a.refreshTree(context.Background())
					ctx.Window.DialogDismiss()
					toastOK(ctx.Window, cat.Saved)
				}),
			},
		}),
	})
}

func openSettings(w *gui.Window) {
	a := appOf(w)
	a.loadSettingsDraft(a.Svc.Config())
	customDialog(w, "settings-dialog", 520, 0, func(win *gui.Window) gui.View {
		return settingsView(appOf(win))
	})
}

func openEnvDialog(w *gui.Window) {
	a := appOf(w)
	if a.SelectedEnvID == 0 {
		return
	}
	a.syncEnvDraft()
	customDialog(w, "env-dialog", 480, 360, func(win *gui.Window) gui.View {
		return envView(appOf(win))
	})
}

func settingsView(app *App) gui.View {
	cat := app.T()
	theme := gui.CurrentTheme()
	fields := []gui.View{
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
	}
	if app.SetProxyMode == config.ProxyManual {
		fields = append(fields, gui.Input(gui.InputCfg{
			ID:    "proxy-url",
			Label: cat.ProxyManual,
			Text:  app.SetProxyURL,
			OnTextChanged: func(s string, ctx gui.EventCtx) {
				appCtx(ctx).SetProxyURL = s
			},
		}))
	}
	fields = append(fields,
		gui.Label(cat.DataDir, theme.TextStyleSecondary),
		gui.InputGroup(gui.InputGroupCfg{
			ID:     "set-datadir-group",
			Sizing: gui.FillFit,
			Segments: []gui.InputGroupSegment{
				gui.InputGroupInput(gui.InputCfg{
					ID:       "data-dir",
					Text:     app.SetDataDir,
					ReadOnly: true,
					Sizing:   gui.FillFit,
				}),
				gui.InputGroupButton(gui.ButtonCfg{
					ID:    "data-dir-browse",
					Label: cat.Browse,
					OnClick: func(ctx gui.EventCtx) {
						pickDataDir(ctx.Window)
					},
				}),
			},
		}),
	)
	dataDirChanged := app.SetDataDir != "" && app.SetDataDir != app.Svc.Config().DataDir
	if dataDirChanged {
		fields = append(fields, gui.Label(cat.ConfirmCopyDataDir, theme.TextStyleSecondary))
	}
	return dialogShell("settings", cat.Settings, []gui.View{
		gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Padding: gui.PaddingMedium,
			Spacing: gui.SpacingSmall,
			Content: fields,
		}),
		gui.Separator(gui.SeparatorCfg{Color: colorDialogRule()}),
		gui.Row(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Padding: gui.PaddingMedium,
			Spacing: gui.SpacingSmall,
			HAlign:  gui.HAlignEnd,
			VAlign:  gui.VAlignMiddle,
			Content: []gui.View{
				gui.TextButtonVariant(cat.Save, gui.ButtonPrimary, func(ctx gui.EventCtx) {
					a := appCtx(ctx)
					copyData := a.SetDataDir != "" && a.SetDataDir != a.Svc.Config().DataDir
					persistSettings(ctx.Window, copyData)
				}),
			},
		}),
	})
}

func pickDataDir(w *gui.Window) {
	a := appOf(w)
	cat := a.T()
	w.NativeFolderDialog(gui.NativeFolderDialogCfg{
		Title:    cat.DataDir,
		StartDir: a.SetDataDir,
		OnDone: func(r gui.NativeDialogResult, win *gui.Window) {
			if r.Status != gui.DialogOK {
				return
			}
			paths := r.PathStrings()
			if len(paths) == 0 {
				return
			}
			appOf(win).SetDataDir = paths[0]
		},
	})
}

func persistSettings(w *gui.Window, copyDataDir bool) {
	a := appOf(w)
	cat := a.T()
	cfg := a.Svc.Config()
	if copyDataDir {
		if err := config.CopyDataDir(cfg.DataDir, a.SetDataDir); err != nil {
			toastErr(w, err.Error())
			return
		}
		cfg.DataDir = a.SetDataDir
	}
	cfg.Language = a.SetLanguage
	cfg.Theme = a.SetTheme
	cfg.Proxy.Mode = a.SetProxyMode
	cfg.Proxy.Manual = a.SetProxyURL
	if err := a.Svc.SaveConfig(); err != nil {
		toastErr(w, err.Error())
		return
	}
	applyWindowTheme(cfg)
	if copyDataDir {
		a.Status = cat.RestartRequired
		toastOK(w, cat.RestartRequired)
	} else {
		toastOK(w, cat.Saved)
	}
	w.DialogDismiss()
}

func trashView(app *App) gui.View {
	cat := app.T()
	cols, err := app.Svc.ListTrashCollections(context.Background())
	if err != nil {
		app.fail(err)
	}
	nodes, err := app.Svc.ListTrashNodes(context.Background())
	if err != nil {
		app.fail(err)
	}
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
						toastErr(ctx.Window, err.Error())
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
						toastErr(ctx.Window, err.Error())
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
				toastErr(ctx.Window, err.Error())
				return
			}
			toastOK(ctx.Window, cat.Saved)
		}),
		gui.TextButton(cat.Cancel, func(ctx gui.EventCtx) {
			a := appCtx(ctx)
			if a.SelectedCollectionID != 0 {
				a.Page = pageCollection
				return
			}
			a.Page = pageHome
		}),
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
	return dialogShell("env", cat.EditEnv, []gui.View{
		gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFill,
			Spacing: gui.SpacingSmall,
			Padding: gui.PaddingMedium,
			Content: []gui.View{
				gui.Input(gui.InputCfg{
					ID:    "env-name",
					Label: cat.Name,
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
				gui.Row(gui.ContainerCfg{
					Sizing:  gui.FillFit,
					Spacing: gui.SpacingSmall,
					HAlign:  gui.HAlignEnd,
					Content: []gui.View{
						gui.TextButtonVariant(cat.Delete, gui.ButtonDanger, func(ctx gui.EventCtx) {
							a := appCtx(ctx)
							if a.SelectedEnvID == 0 {
								return
							}
							if err := a.Svc.DeleteEnvironment(context.Background(), a.SelectedEnvID); err != nil {
								toastErr(ctx.Window, err.Error())
								return
							}
							a.SelectedEnvID = 0
							a.refreshTree(context.Background())
							if len(a.Envs) > 0 {
								a.SelectedEnvID = a.Envs[0].ID
							}
							a.syncEnvDraft()
							ctx.Window.DialogDismiss()
						}),
						gui.TextButtonVariant(cat.Save, gui.ButtonPrimary, func(ctx gui.EventCtx) {
							a := appCtx(ctx)
							if a.SelectedEnvID == 0 {
								return
							}
							e, err := a.Svc.GetEnvironment(context.Background(), a.SelectedEnvID)
							if err != nil {
								toastErr(ctx.Window, err.Error())
								return
							}
							e.Name = a.DraftEnvName
							if err := a.Svc.UpdateEnvironment(context.Background(), e, decodeMap(a.DraftEnvVars)); err != nil {
								toastErr(ctx.Window, err.Error())
								return
							}
							a.refreshTree(context.Background())
							ctx.Window.DialogDismiss()
							toastOK(ctx.Window, cat.Saved)
						}),
					},
				}),
			},
		}),
	})
}
