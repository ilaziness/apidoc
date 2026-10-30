package ui

import (
	"github.com/go-gui-org/go-gui/gui"
)

// customDialog opens a modal with product chrome (PaddingNone + body fill).
func customDialog(w *gui.Window, id string, width, height float32, view func(*gui.Window) gui.View) {
	cfg := gui.DialogCfg{
		ID:         id,
		DialogType: gui.DialogCustom,
		Width:      width,
		Padding:    gui.PaddingNone,
		Radius:     gui.RadiusMedium,
		Color:      colorDialogBody(),
		CustomView: view,
	}
	if height > 0 {
		cfg.Height = height
	}
	w.Dialog(cfg)
}

// dialogShell wraps custom-dialog content with the product chrome:
// title bar (title + × close) then body. See PRD §7.
func dialogShell(id, title string, body []gui.View) gui.View {
	content := make([]gui.View, 0, 2+len(body))
	content = append(content,
		dialogTitleBar(id, title),
		gui.Separator(gui.SeparatorCfg{Color: colorDialogRule()}),
	)
	content = append(content, body...)
	return gui.Column(gui.ContainerCfg{
		Sizing:       gui.FillFill,
		Spacing:      gui.SpacingPx(0),
		Padding:      gui.PaddingNone,
		Radius:       gui.RadiusMedium,
		SizeBorder:   gui.NoBorder,
		Clip:         true,
		ClipContents: true,
		Color:        colorDialogBody(),
		Content:      content,
	})
}

func dialogTitleBar(id, title string) gui.View {
	return gui.ViewFunc(func(w *gui.Window) gui.View {
		theme := gui.CurrentTheme()
		closeTip := appOf(w).T().Close
		return gui.Row(gui.ContainerCfg{
			Sizing:     gui.FillFit,
			Padding:    gui.NewPadding(4, 6, 4, 12),
			Radius:     gui.NoRadius,
			SizeBorder: gui.NoBorder,
			Color:      colorDialogTitle(),
			VAlign:     gui.VAlignMiddle,
			Spacing:    gui.SpacingPx(0),
			Content: []gui.View{
				gui.Label(title, theme.TextStyleTitleSmall),
				gui.Column(gui.ContainerCfg{Sizing: gui.FillFit}),
				gui.Button(gui.ButtonCfg{
					ID:      id + "-close",
					Variant: gui.ButtonGhost,
					Padding: gui.NewPadding(4, 6, 4, 6),
					Radius:  gui.RadiusSmall,
					A11YCfg: gui.A11YCfg{A11YLabel: closeTip},
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: gui.IconClose, TextStyle: theme.TextStyleIconMedium}),
					},
					OnClick: func(ctx gui.EventCtx) {
						ctx.Window.DialogDismiss()
					},
				}),
			},
		})
	})
}
