package ui

import (
	"context"

	"github.com/go-gui-org/go-gui/gui"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

func editorView(w *gui.Window, app *App) gui.View {
	cat := app.T()
	if app.SelectedNodeID == 0 {
		return paddedHint(cat.NoSelection)
	}
	switch app.selectedKind() {
	case store.NodeHTTP:
		return httpWorkspace(app)
	case store.NodeMarkdown:
		return markdownWorkspace(w, app)
	case store.NodeFolder:
		return folderWorkspace(app)
	default:
		return paddedHint(cat.NoSelection)
	}
}

func editorNameRow(app *App) gui.View {
	cat := app.T()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SpacingSmall,
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Input(gui.InputCfg{
				ID:     "node-name",
				Label:  cat.Rename,
				Text:   app.DraftName,
				Sizing: gui.FillFit,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					a := appCtx(ctx)
					a.DraftName = s
					a.Dirty = true
				},
			}),
			gui.TextButton(cat.Save, func(ctx gui.EventCtx) { _ = saveCurrent(ctx.Window) }),
			gui.TextButtonVariant(cat.Discard, gui.ButtonGhost, func(ctx gui.EventCtx) {
				appCtx(ctx).discardDraft(context.Background())
			}),
			gui.TextButtonVariant(cat.Delete, gui.ButtonDanger, func(ctx gui.EventCtx) {
				confirmDeleteNode(ctx.Window)
			}),
		},
	})
}

func saveCurrent(w *gui.Window) error {
	a := appOf(w)
	if a.SelectedNodeID == 0 {
		return nil
	}
	n, err := a.Svc.GetNode(context.Background(), a.SelectedNodeID)
	if err != nil {
		a.fail(err)
		return err
	}
	n.Name = a.DraftName
	switch n.Kind {
	case store.NodeMarkdown:
		n.Markdown = a.DraftMarkdown
	case store.NodeHTTP:
		raw, err := httpapi.EncodeSpec(a.DraftHTTP)
		if err != nil {
			a.fail(err)
			return err
		}
		n.HTTP = raw
	}
	if err := a.Svc.SaveNode(context.Background(), n); err != nil {
		a.fail(err)
		return err
	}
	a.Dirty = false
	a.refreshTree(context.Background())
	toastOK(w, a.T().Saved)
	return nil
}

func markdownWorkspace(w *gui.Window, app *App) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Padding: gui.PaddingSmall,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			editorNameRow(app),
			gui.TabControl(gui.TabControlCfg{
				ID:       "md-work-tabs",
				Sizing:   gui.FillFill,
				Selected: mdWorkTab(app.WorkTab),
				OnSelect: func(id string, ctx gui.EventCtx) {
					appCtx(ctx).WorkTab = id
				},
				Items: []gui.TabItemCfg{
					gui.NewTabItem(workPreview, cat.Preview, []gui.View{
						w.Markdown(gui.MarkdownCfg{Source: app.DraftMarkdown}),
					}),
					gui.NewTabItem(workEdit, cat.EditDoc, []gui.View{
						gui.Input(gui.InputCfg{
							ID:     "md-body",
							Label:  cat.Markdown,
							Mode:   gui.InputMultiline,
							Text:   app.DraftMarkdown,
							Sizing: gui.FillFill,
							OnTextChanged: func(s string, ctx gui.EventCtx) {
								a := appCtx(ctx)
								a.DraftMarkdown = s
								a.Dirty = true
							},
						}),
					}),
				},
			}),
		},
	})
}

func mdWorkTab(tab string) string {
	if tab == workEdit {
		return workEdit
	}
	return workPreview
}

func folderWorkspace(app *App) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Padding: gui.PaddingSmall,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{editorNameRow(app)},
	})
}
