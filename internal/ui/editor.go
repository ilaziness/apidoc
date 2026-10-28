package ui

import (
	"context"

	"github.com/go-gui-org/go-gui/gui"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

func editorView(w *gui.Window, app *App) gui.View {
	cat := app.T()
	if app.SelectedCollectionID == 0 {
		return paddedHint(cat.EmptyHint)
	}
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
			gui.TextButton(cat.Save, func(ctx gui.EventCtx) { saveCurrent(ctx.Window) }),
			gui.TextButtonVariant(cat.Discard, gui.ButtonGhost, func(ctx gui.EventCtx) {
				appCtx(ctx).discardDraft(context.Background())
			}),
			gui.TextButtonVariant(cat.Delete, gui.ButtonDanger, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				if err := a.Svc.DeleteNode(context.Background(), a.SelectedNodeID); err != nil {
					a.fail(err)
					return
				}
				a.SelectedNodeID = 0
				a.Dirty = false
				a.refreshTree(context.Background())
			}),
		},
	})
}

func saveCurrent(w *gui.Window) {
	a := appOf(w)
	if a.SelectedNodeID == 0 {
		return
	}
	n, err := a.Svc.GetNode(context.Background(), a.SelectedNodeID)
	if err != nil {
		a.fail(err)
		return
	}
	n.Name = a.DraftName
	switch n.Kind {
	case store.NodeMarkdown:
		n.Markdown = a.DraftMarkdown
	case store.NodeHTTP:
		raw, err := httpapi.EncodeSpec(a.DraftHTTP)
		if err != nil {
			a.fail(err)
			return
		}
		n.HTTP = raw
	}
	if err := a.Svc.SaveNode(context.Background(), n); err != nil {
		a.fail(err)
		return
	}
	a.Dirty = false
	a.refreshTree(context.Background())
	a.Status = a.T().Saved
	toastOK(w, a.T().Saved)
}

func markdownWorkspace(w *gui.Window, app *App) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			editorNameRow(app),
			gui.Splitter(gui.SplitterCfg{
				ID:          "md-split",
				Sizing:      gui.FillFill,
				Orientation: gui.SplitterHorizontal,
				Ratio:       gui.SomeF(0.5),
				First: gui.SplitterPaneCfg{
					MinSize: 160,
					Content: []gui.View{
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
					},
				},
				Second: gui.SplitterPaneCfg{
					MinSize: 160,
					Content: []gui.View{
						gui.Label(cat.Preview, gui.CurrentTheme().TextStyleTitleSmall),
						w.Markdown(gui.MarkdownCfg{Source: app.DraftMarkdown}),
					},
				},
			}),
		},
	})
}

func folderWorkspace(app *App) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{editorNameRow(app)},
	})
}
