package ui

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-gui-org/go-gui/gui"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

func sidebar(w *gui.Window, app *App) gui.View {
	cat := app.T()
	theme := gui.CurrentTheme()
	head := []gui.View{
		gui.Input(gui.InputCfg{
			ID:    "collection-rename",
			Label: cat.Rename,
			Text:  app.selectedCollectionName(),
			OnTextChanged: func(s string, ctx gui.EventCtx) {
				a := appCtx(ctx)
				if a.SelectedCollectionID == 0 || strings.TrimSpace(s) == "" {
					return
				}
				if err := a.Svc.RenameCollection(context.Background(), a.SelectedCollectionID, s); err != nil {
					toastErr(ctx.Window, err.Error())
					return
				}
				for i := range a.Collections {
					if a.Collections[i].ID == a.SelectedCollectionID {
						a.Collections[i].Name = s
						break
					}
				}
			},
		}),
		gui.Row(gui.ContainerCfg{
			Spacing: gui.SpacingSmall,
			Content: []gui.View{
				gui.TextButtonVariant("+ "+cat.Folder, gui.ButtonGhost, func(ctx gui.EventCtx) {
					createUnder(ctx, store.NodeFolder)
				}),
				gui.TextButtonVariant("+ HTTP", gui.ButtonGhost, func(ctx gui.EventCtx) {
					createUnder(ctx, store.NodeHTTP)
				}),
				gui.TextButtonVariant("+ MD", gui.ButtonGhost, func(ctx gui.EventCtx) {
					createUnder(ctx, store.NodeMarkdown)
				}),
			},
		}),
	}
	if app.SelectedCollectionID == 0 {
		head = []gui.View{gui.Label(cat.EmptyHint, theme.TextStyleSecondary)}
	}
	content := append(head, gui.Tree(gui.TreeCfg{
		ID:     "nav-tree",
		Sizing: gui.FillFill,
		Nodes:  buildTreeNodes(app),
		OnSelect: func(id string, ctx gui.EventCtx) {
			nid, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return
			}
			askUnsaved(ctx.Window, func() {
				appOf(ctx.Window).loadNode(context.Background(), nid)
			})
		},
	}))
	col := gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Padding: gui.PaddingSmall,
		Color:   theme.ColorPanel,
		Content: content,
	})
	return gui.ContextMenu(w, gui.ContextMenuCfg{
		ID:     "tree-ctx",
		Sizing: gui.FillFill,
		Items: []gui.MenuItemCfg{
			gui.MenuItemText("ctx-folder", "+ "+cat.Folder),
			gui.MenuItemText("ctx-http", "+ HTTP"),
			gui.MenuItemText("ctx-md", "+ MD"),
			gui.MenuSeparator(),
			gui.MenuItemText("ctx-delete", cat.Delete),
		},
		Action: func(id string, ctx gui.EventCtx) {
			switch id {
			case "ctx-folder":
				createUnder(ctx, store.NodeFolder)
			case "ctx-http":
				createUnder(ctx, store.NodeHTTP)
			case "ctx-md":
				createUnder(ctx, store.NodeMarkdown)
			case "ctx-delete":
				a := appCtx(ctx)
				if a.SelectedNodeID == 0 {
					return
				}
				if err := a.Svc.DeleteNode(context.Background(), a.SelectedNodeID); err != nil {
					toastErr(ctx.Window, err.Error())
					return
				}
				a.SelectedNodeID = 0
				a.Dirty = false
				a.refreshTree(context.Background())
			}
		},
		Content: []gui.View{col},
	})
}

func createUnder(ctx gui.EventCtx, kind string) {
	a := appCtx(ctx)
	if a.SelectedCollectionID == 0 {
		return
	}
	parent := a.insertParent()
	cat := a.T()
	var err error
	switch kind {
	case store.NodeFolder:
		_, err = a.Svc.CreateFolder(context.Background(), a.SelectedCollectionID, parent, cat.Folder)
	case store.NodeHTTP:
		_, err = a.Svc.CreateHTTP(context.Background(), a.SelectedCollectionID, parent, cat.HTTPRequest, httpapi.DefaultSpec())
	case store.NodeMarkdown:
		_, err = a.Svc.CreateMarkdown(context.Background(), a.SelectedCollectionID, parent, cat.Markdown, "")
	}
	if err != nil {
		a.fail(err)
		return
	}
	a.refreshTree(context.Background())
}

func buildTreeNodes(app *App) []gui.TreeNodeCfg {
	byParent := map[int64][]store.Node{}
	var roots []store.Node
	for _, n := range app.Nodes {
		if n.ParentID == nil {
			roots = append(roots, n)
			continue
		}
		byParent[*n.ParentID] = append(byParent[*n.ParentID], n)
	}
	var walk func(n store.Node) gui.TreeNodeCfg
	walk = func(n store.Node) gui.TreeNodeCfg {
		label := treeLabel(n)
		children := byParent[n.ID]
		nodes := make([]gui.TreeNodeCfg, 0, len(children))
		for _, c := range children {
			nodes = append(nodes, walk(c))
		}
		return gui.TreeNodeCfg{
			ID:    strconv.FormatInt(n.ID, 10),
			Text:  label,
			Nodes: nodes,
		}
	}
	out := make([]gui.TreeNodeCfg, 0, len(roots))
	for _, r := range roots {
		out = append(out, walk(r))
	}
	return out
}

func treeLabel(n store.Node) string {
	switch n.Kind {
	case store.NodeFolder:
		return "▸ " + n.Name
	case store.NodeMarkdown:
		return "MD  " + n.Name
	case store.NodeHTTP:
		method := "GET"
		if spec, err := httpapi.DecodeSpec(n.HTTP); err == nil && spec.Method != "" {
			method = spec.Method
		}
		return method + "  " + n.Name
	default:
		return n.Name
	}
}
