package ui

import (
	"context"

	"github.com/go-gui-org/go-gui/gui"
)

func mainView(w *gui.Window) gui.View {
	app := appOf(w)
	theme := gui.CurrentTheme()
	body := []gui.View{
		toolbar(w, app),
		bodyView(w, app),
	}
	if app.Status != "" {
		body = append(body, gui.Label(app.Status, theme.TextStyleSecondary))
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Color:   theme.ColorBackground,
		Padding: gui.PaddingNone,
		Spacing: gui.SpacingPx(0),
		Content: body,
	})
}

func toolbar(w *gui.Window, app *App) gui.View {
	cat := app.T()
	theme := gui.CurrentTheme()
	if !app.toolbarTipArmed {
		app.toolbarTip = ""
	}
	app.toolbarTipArmed = false
	items := []gui.View{
		iconBtn("tb-home", gui.IconHome, cat.Home, func(ctx gui.EventCtx) {
			askUnsaved(ctx.Window, func() { appOf(ctx.Window).goHome() })
		}),
		iconBtn("tb-new-col", gui.IconPlus, cat.NewCollection, func(ctx gui.EventCtx) {
			openNewProjectDialog(ctx.Window)
		}),
		iconBtn("tb-import", gui.IconUpload, cat.ImportOpenAPI, func(ctx gui.EventCtx) {
			openImport(ctx.Window)
		}),
	}
	if app.Page == pageCollection {
		envOpts := make([]gui.SelectOption, 0, len(app.Envs))
		for _, e := range app.Envs {
			envOpts = append(envOpts, gui.NewSelectOption(e.Name, idString(e.ID)))
		}
		items = append(items,
			gui.Select(gui.SelectCfg{
				ID:          "tb-env",
				Placeholder: cat.Environments,
				Options:     envOpts,
				Selected:    selectedOrEmpty(idString(app.SelectedEnvID)),
				MinWidth:    140,
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
			iconBtn("tb-new-env", gui.IconPlus, cat.NewEnvironment, func(ctx gui.EventCtx) {
				createEnvironment(ctx.Window)
			}),
			iconBtn("tb-edit-env", gui.IconEdit, cat.EditEnv, func(ctx gui.EventCtx) {
				openEnvDialog(ctx.Window)
			}),
			iconBtn("tb-del-col", gui.IconClose, cat.DeleteCollection, func(ctx gui.EventCtx) {
				confirmDeleteCollection(ctx.Window)
			}),
		)
	}
	items = append(items, gui.Column(gui.ContainerCfg{Sizing: gui.FillFit}))
	if app.toolbarTip != "" {
		items = append(items, gui.Label(app.toolbarTip, theme.TextStyleSecondary))
	}
	items = append(items,
		iconBtn("tb-trash", gui.IconTrash, cat.Trash, func(ctx gui.EventCtx) {
			askUnsaved(ctx.Window, func() {
				a := appOf(ctx.Window)
				a.stopSend()
				a.Page = pageTrash
			})
		}),
		iconBtn("tb-settings", gui.IconWrench, cat.Settings, func(ctx gui.EventCtx) {
			openSettings(ctx.Window)
		}),
	)
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NewPadding(2, 4, 2, 4),
		Spacing: gui.SpacingPx(2),
		VAlign:  gui.VAlignMiddle,
		Color:   theme.ColorPanel,
		Content: items,
	})
}

func iconBtn(id, icon, tip string, onClick func(gui.EventCtx)) gui.View {
	theme := gui.CurrentTheme()
	return gui.Button(gui.ButtonCfg{
		ID:      id,
		Variant: gui.ButtonGhost,
		Padding: gui.NewPadding(4, 6, 4, 6),
		Radius:  gui.RadiusSmall,
		A11YCfg: gui.A11YCfg{A11YDescription: tip},
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: icon, TextStyle: theme.TextStyleIconXLarge}),
		},
		OnClick: onClick,
		OnHover: func(ctx gui.EventCtx) {
			a := appCtx(ctx)
			a.toolbarTip = tip
			a.toolbarTipArmed = true
		},
	})
}

func bodyView(w *gui.Window, app *App) gui.View {
	switch app.Page {
	case pageTrash:
		return trashView(app)
	case pageCollection:
		return collectionView(w, app)
	default:
		return homeView(app)
	}
}

func collectionView(w *gui.Window, app *App) gui.View {
	return gui.Splitter(gui.SplitterCfg{
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
			Content: []gui.View{editorView(w, app)},
		},
	})
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
	customDialog(w, "unsaved-dialog", 420, 0, func(win *gui.Window) gui.View {
		c := appOf(win).T()
		return dialogShell("unsaved", c.UnsavedPrompt, []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.PaddingMedium,
				Spacing: gui.SpacingSmall,
				HAlign:  gui.HAlignEnd,
				Content: []gui.View{
					gui.TextButtonVariant(c.Save, gui.ButtonPrimary, func(ctx gui.EventCtx) {
						if err := saveCurrent(ctx.Window); err != nil {
							return
						}
						ctx.Window.DialogDismiss()
						next()
					}),
					gui.TextButton(c.Discard, func(ctx gui.EventCtx) {
						appOf(ctx.Window).discardDraft(context.Background())
						ctx.Window.DialogDismiss()
						next()
					}),
				},
			}),
		})
	})
}

func createEnvironment(w *gui.Window) {
	a := appOf(w)
	e, err := a.Svc.CreateEnvironment(context.Background(), a.SelectedCollectionID, "dev", map[string]string{"baseUrl": "http://localhost"})
	if err != nil {
		toastErr(w, err.Error())
		return
	}
	a.refreshTree(context.Background())
	a.SelectedEnvID = e.ID
	a.syncEnvDraft()
	a.Svc.Config().LastEnvironmentID = e.ID
	_ = a.Svc.SaveConfig()
	openEnvDialog(w)
}

func confirmDeleteCollection(w *gui.Window) {
	a := appOf(w)
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
			app.goHome()
			app.refreshCollections(context.Background())
			toastOK(win, cat.Saved)
		},
	})
}

func confirmDeleteNode(w *gui.Window) {
	a := appOf(w)
	if a.SelectedNodeID == 0 {
		return
	}
	cat := a.T()
	id := a.SelectedNodeID
	w.Dialog(gui.DialogCfg{
		ID:            "delete-node",
		DialogType:    gui.DialogConfirm,
		Title:         cat.Delete,
		Body:          cat.ConfirmDelete,
		DefaultButton: gui.DialogButtonNo,
		OnOkYes: func(win *gui.Window) {
			app := appOf(win)
			if err := app.Svc.DeleteNode(context.Background(), id); err != nil {
				toastErr(win, err.Error())
				return
			}
			app.SelectedNodeID = 0
			app.Dirty = false
			app.refreshTree(context.Background())
		},
	})
}

func toastOK(w *gui.Window, title string) {
	w.Toast(gui.ToastCfg{Title: title, Severity: gui.ToastSuccess})
}

func toastErr(w *gui.Window, title string) {
	w.Toast(gui.ToastCfg{Title: title, Severity: gui.ToastError})
}
