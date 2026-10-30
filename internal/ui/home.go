package ui

import (
	"context"
	"strconv"

	"github.com/go-gui-org/go-gui/gui"
)

func homeView(app *App) gui.View {
	cat := app.T()
	theme := gui.CurrentTheme()
	tiles := make([]gui.View, 0, len(app.Collections)+1)
	tiles = append(tiles, collectionTile("home-new", gui.IconPlus, cat.NewCollection, theme.TextStyleSecondary, func(ctx gui.EventCtx) {
		openNewProjectDialog(ctx.Window)
	}))
	for _, c := range app.Collections {
		id := c.ID
		name := c.Name
		tiles = append(tiles, collectionTile("home-col-"+strconv.FormatInt(id, 10), gui.IconFolder, name, theme.TextStyleBody, func(ctx gui.EventCtx) {
			askUnsaved(ctx.Window, func() {
				appOf(ctx.Window).selectCollection(context.Background(), id)
			})
		}))
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Padding: gui.PaddingMedium,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Label(cat.Collections, theme.TextStyleTitle),
			gui.Wrap(gui.ContainerCfg{
				Sizing:  gui.FillFill,
				Spacing: gui.SpacingMedium,
				Content: tiles,
			}),
		},
	})
}

func collectionTile(id, icon, title string, titleStyle gui.TextStyle, onClick func(gui.EventCtx)) gui.View {
	theme := gui.CurrentTheme()
	return gui.Button(gui.ButtonCfg{
		ID:      id,
		Width:   220,
		Height:  110,
		Variant: gui.ButtonSecondary,
		OnClick: onClick,
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FillFill,
				Spacing: gui.SpacingSmall,
				HAlign:  gui.HAlignCenter,
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: icon, TextStyle: theme.TextStyleIconMedium}),
					gui.Label(title, titleStyle),
				},
			}),
		},
	})
}
