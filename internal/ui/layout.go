package ui

import (
	"context"

	"github.com/go-gui-org/go-gui/gui"
)

func mainView(w *gui.Window) gui.View {
	app := appOf(w)
	theme := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing: gui.FillFill,
		Color:  theme.ColorBackground,
		Content: []gui.View{
			gui.Menubar(w, gui.MenubarCfg{
				ID:     "menubar",
				Sizing: gui.FillFit,
				Action: menuAction,
				Items:  menuItems(app),
			}),
			chrome(app),
			gui.Separator(gui.SeparatorCfg{}),
			gui.Splitter(gui.SplitterCfg{
				ID:          "main-split",
				Sizing:      gui.FillFill,
				Orientation: gui.SplitterHorizontal,
				Ratio:       gui.SomeF(app.SideRatio),
				OnChange: func(ratio float32, _ gui.SplitterCollapsed, ctx gui.EventCtx) {
					appCtx(ctx).SideRatio = ratio
				},
				First: gui.SplitterPaneCfg{
					MinSize:     200,
					Collapsible: true,
					Content:     []gui.View{sidebar(w, app)},
				},
				Second: gui.SplitterPaneCfg{
					MinSize: 480,
					Content: []gui.View{mainPanel(w, app)},
				},
			}),
			gui.Label(app.Status, theme.TextStyleSecondary),
		},
	})
}

func menuItems(app *App) []gui.MenuItemCfg {
	cat := app.T()
	return []gui.MenuItemCfg{
		gui.MenuSubmenu("menu-collection", cat.Collections, []gui.MenuItemCfg{
			gui.MenuItemText("menu-new-col", cat.NewCollection),
			gui.MenuItemText("menu-import", cat.ImportOpenAPI),
			gui.MenuSeparator(),
			gui.MenuItemText("menu-delete-col", cat.DeleteCollection),
		}),
		gui.MenuSubmenu("menu-env", cat.Environments, []gui.MenuItemCfg{
			gui.MenuItemText("menu-new-env", cat.NewEnvironment),
			gui.MenuItemText("menu-edit-env", cat.EditEnv),
		}),
		gui.MenuItemText("menu-trash", cat.Trash),
		gui.MenuItemText("menu-settings", cat.Settings),
	}
}

func menuAction(id string, ctx gui.EventCtx) {
	w := ctx.Window
	a := appOf(w)
	switch id {
	case "menu-new-col":
		askUnsaved(w, func() { createCollection(w) })
	case "menu-import":
		openImport(w)
	case "menu-delete-col":
		confirmDeleteCollection(w)
	case "menu-new-env":
		createEnvironment(w)
	case "menu-edit-env":
		if a.SelectedEnvID == 0 {
			return
		}
		a.syncEnvDraft()
		a.Panel = panelEnv
	case "menu-trash":
		a.Panel = panelTrash
	case "menu-settings":
		a.Panel = panelSettings
	}
}

func chrome(app *App) gui.View {
	cat := app.T()
	colOpts := make([]gui.SelectOption, 0, len(app.Collections))
	for _, c := range app.Collections {
		colOpts = append(colOpts, gui.NewSelectOption(c.Name, idString(c.ID)))
	}
	envOpts := make([]gui.SelectOption, 0, len(app.Envs))
	for _, e := range app.Envs {
		envOpts = append(envOpts, gui.NewSelectOption(e.Name, idString(e.ID)))
	}
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.PaddingSmall,
		Spacing: gui.SpacingSmall,
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Select(gui.SelectCfg{
				ID:          "chrome-collection",
				Placeholder: cat.Collections,
				Options:     colOpts,
				Selected:    selectedOrEmpty(idString(app.SelectedCollectionID)),
				MinWidth:    200,
				OnSelect: func(vals []string, ctx gui.EventCtx) {
					id, ok := parseID(vals)
					if !ok {
						return
					}
					askUnsaved(ctx.Window, func() {
						appOf(ctx.Window).selectCollection(context.Background(), id)
					})
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				HAlign:  gui.HAlignEnd,
				Spacing: gui.SpacingSmall,
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{
					gui.Select(gui.SelectCfg{
						ID:          "chrome-env",
						Placeholder: cat.Environments,
						Options:     envOpts,
						Selected:    selectedOrEmpty(idString(app.SelectedEnvID)),
						MinWidth:    160,
						Disabled:    app.SelectedCollectionID == 0,
						OnSelect: func(vals []string, ctx gui.EventCtx) {
							id, ok := parseID(vals)
							if !ok {
								return
							}
							a := appCtx(ctx)
							a.SelectedEnvID = id
							a.syncEnvDraft()
							a.Svc.Config().LastEnvironmentID = id
							_ = a.Svc.SaveConfig()
						},
					}),
				},
			}),
		},
	})
}

func mainPanel(w *gui.Window, app *App) gui.View {
	switch app.Panel {
	case panelSettings:
		return settingsView(app)
	case panelTrash:
		return trashView(app)
	case panelImport:
		return importView(app)
	case panelEnv:
		return envView(app)
	default:
		return editorView(w, app)
	}
}

func closeEditor(ctx gui.EventCtx) {
	appCtx(ctx).Panel = panelEditor
}

func paddedHint(text string) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Padding: gui.PaddingLarge,
		Content: []gui.View{gui.Label(text, gui.CurrentTheme().TextStyleSecondary)},
	})
}

func askUnsaved(w *gui.Window, next func()) {
	a := appOf(w)
	if !a.Dirty {
		next()
		return
	}
	cat := a.T()
	w.Dialog(gui.DialogCfg{
		ID:         "unsaved-dialog",
		DialogType: gui.DialogCustom,
		Title:      cat.UnsavedPrompt,
		CustomView: func(win *gui.Window) gui.View {
			c := appOf(win).T()
			return gui.Row(gui.ContainerCfg{
				Spacing: gui.SpacingSmall,
				Content: []gui.View{
					gui.TextButtonVariant(c.Save, gui.ButtonPrimary, func(ctx gui.EventCtx) {
						saveCurrent(ctx.Window)
						ctx.Window.DialogDismiss()
						next()
					}),
					gui.TextButton(c.Discard, func(ctx gui.EventCtx) {
						appOf(ctx.Window).discardDraft(context.Background())
						ctx.Window.DialogDismiss()
						next()
					}),
					gui.TextButton(c.Cancel, func(ctx gui.EventCtx) {
						ctx.Window.DialogDismiss()
					}),
				},
			})
		},
	})
}

func createCollection(w *gui.Window) {
	a := appOf(w)
	cat := a.T()
	c, err := a.Svc.CreateCollection(context.Background(), cat.NewCollection)
	if err != nil {
		toastErr(w, err.Error())
		return
	}
	a.refreshCollections(context.Background())
	a.selectCollection(context.Background(), c.ID)
}

func createEnvironment(w *gui.Window) {
	a := appOf(w)
	if a.SelectedCollectionID == 0 {
		return
	}
	e, err := a.Svc.CreateEnvironment(context.Background(), a.SelectedCollectionID, "dev", map[string]string{"baseUrl": "http://localhost"})
	if err != nil {
		toastErr(w, err.Error())
		return
	}
	a.refreshTree(context.Background())
	a.SelectedEnvID = e.ID
	a.syncEnvDraft()
	a.Panel = panelEnv
}

func confirmDeleteCollection(w *gui.Window) {
	a := appOf(w)
	if a.SelectedCollectionID == 0 {
		return
	}
	cat := a.T()
	id := a.SelectedCollectionID
	w.Dialog(gui.DialogCfg{
		ID:            "delete-col",
		DialogType:    gui.DialogConfirm,
		Title:         cat.DeleteCollection,
		Body:          cat.ConfirmDelete,
		DefaultButton: gui.DialogButtonNo,
		OnOkYes: func(win *gui.Window) {
			app := appOf(win)
			if err := app.Svc.DeleteCollection(context.Background(), id); err != nil {
				toastErr(win, err.Error())
				return
			}
			app.SelectedCollectionID = 0
			app.SelectedNodeID = 0
			app.Dirty = false
			app.Nodes = nil
			app.Envs = nil
			app.refreshCollections(context.Background())
			app.Status = ""
			toastOK(win, cat.Saved)
		},
	})
}

func openImport(w *gui.Window) {
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
			a.Panel = panelImport
		},
	})
}

func toastOK(w *gui.Window, title string) {
	w.Toast(gui.ToastCfg{Title: title, Severity: gui.ToastSuccess})
}

func toastErr(w *gui.Window, title string) {
	w.Toast(gui.ToastCfg{Title: title, Severity: gui.ToastError})
}
