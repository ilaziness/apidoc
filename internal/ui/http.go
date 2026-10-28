package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-gui-org/go-gui/gui"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

func httpWorkspace(app *App) gui.View {
	cat := app.T()
	spec := app.DraftHTTP
	sendLabel := cat.Send
	if app.isSending() {
		sendLabel = cat.Cancel
	}
	urlBar := gui.InputGroup(gui.InputGroupCfg{
		ID:     "http-urlbar",
		Sizing: gui.FillFit,
		Segments: []gui.InputGroupSegment{
			gui.InputGroupSelect(gui.SelectCfg{
				ID:       "http-method",
				MinWidth: 96,
				Items:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
				Selected: []string{spec.Method},
				OnSelect: func(vals []string, ctx gui.EventCtx) {
					if len(vals) == 0 {
						return
					}
					a := appCtx(ctx)
					a.DraftHTTP.Method = vals[0]
					a.Dirty = true
				},
			}),
			gui.InputGroupInput(gui.InputCfg{
				ID:     "http-url",
				Text:   spec.URL,
				Sizing: gui.FillFit,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					a := appCtx(ctx)
					a.DraftHTTP.URL = s
					a.Dirty = true
				},
			}),
			gui.InputGroupButton(gui.ButtonCfg{
				ID:      "http-send",
				Label:   sendLabel,
				Variant: gui.ButtonPrimary,
				OnClick: func(ctx gui.EventCtx) {
					if appCtx(ctx).isSending() {
						cancelSend(ctx.Window)
						return
					}
					sendCurrent(ctx.Window)
				},
			}),
		},
	})

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			editorNameRow(app),
			urlBar,
			gui.Splitter(gui.SplitterCfg{
				ID:          "http-split",
				Sizing:      gui.FillFill,
				Orientation: gui.SplitterVertical,
				Ratio:       gui.SomeF(app.ReqRatio),
				OnChange: func(ratio float32, _ gui.SplitterCollapsed, ctx gui.EventCtx) {
					appCtx(ctx).ReqRatio = ratio
				},
				First: gui.SplitterPaneCfg{
					MinSize: 160,
					Content: []gui.View{httpTabs(app)},
				},
				Second: gui.SplitterPaneCfg{
					MinSize: 120,
					Content: []gui.View{
						gui.Column(gui.ContainerCfg{
							Sizing:  gui.FillFill,
							Spacing: gui.SpacingSmall,
							Content: []gui.View{
								gui.Label(cat.Response, gui.CurrentTheme().TextStyleTitleSmall),
								gui.Input(gui.InputCfg{
									ID:       "resp",
									Mode:     gui.InputMultiline,
									ReadOnly: true,
									Text:     app.RespText,
									Sizing:   gui.FillFill,
								}),
							},
						}),
					},
				},
			}),
		},
	})
}

func httpTabs(app *App) gui.View {
	cat := app.T()
	return gui.TabControl(gui.TabControlCfg{
		ID:       "http-tabs",
		Sizing:   gui.FillFill,
		Selected: app.EditorTab,
		OnSelect: func(id string, ctx gui.EventCtx) {
			appCtx(ctx).EditorTab = id
		},
		Items: []gui.TabItemCfg{
			gui.NewTabItem(tabParams, cat.Params, []gui.View{kvPairEditors(app)}),
			gui.NewTabItem(tabHeaders, cat.Headers, []gui.View{
				kvEditor("http-headers", cat.Headers, app.DraftHTTP.Headers, func(a *App, rows []httpapi.KV) {
					a.DraftHTTP.Headers = rows
				}),
			}),
			gui.NewTabItem(tabBody, cat.Body, []gui.View{bodyEditor(app)}),
			gui.NewTabItem(tabAuth, cat.Auth, []gui.View{authEditor(app)}),
			gui.NewTabItem(tabOptions, cat.Options, []gui.View{optionsEditor(app)}),
		},
	})
}

func kvPairEditors(app *App) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			kvEditor("http-query", cat.Query, app.DraftHTTP.Query, func(a *App, rows []httpapi.KV) {
				a.DraftHTTP.Query = rows
			}),
			kvEditor("http-path", cat.PathParams, app.DraftHTTP.PathParams, func(a *App, rows []httpapi.KV) {
				a.DraftHTTP.PathParams = rows
			}),
		},
	})
}

func kvEditor(id, label string, rows []httpapi.KV, apply func(*App, []httpapi.KV)) gui.View {
	return gui.Input(gui.InputCfg{
		ID:     id,
		Label:  label + "  (Key:Value, prefix - to skip)",
		Mode:   gui.InputMultiline,
		Text:   encodeKV(rows),
		Sizing: gui.FillFill,
		OnTextChanged: func(s string, ctx gui.EventCtx) {
			a := appCtx(ctx)
			apply(a, decodeKV(s))
			a.Dirty = true
		},
	})
}

func bodyEditor(app *App) gui.View {
	kind := app.DraftHTTP.Body.Kind
	if kind == "" {
		kind = "none"
	}
	body := app.DraftHTTP.Body
	fields := []gui.View{
		gui.Select(gui.SelectCfg{
			ID:       "body-kind",
			Label:    app.T().Body,
			Items:    []string{"none", "json", "form", "multipart", "raw", "binary"},
			Selected: []string{kind},
			OnSelect: func(vals []string, ctx gui.EventCtx) {
				if len(vals) == 0 {
					return
				}
				a := appCtx(ctx)
				a.DraftHTTP.Body.Kind = vals[0]
				a.Dirty = true
			},
		}),
	}
	switch kind {
	case "json":
		fields = append(fields, gui.Input(gui.InputCfg{
			ID:     "body-json",
			Mode:   gui.InputMultiline,
			Text:   body.JSON,
			Sizing: gui.FillFill,
			OnTextChanged: func(s string, ctx gui.EventCtx) {
				a := appCtx(ctx)
				a.DraftHTTP.Body.Kind = "json"
				a.DraftHTTP.Body.JSON = s
				a.Dirty = true
			},
		}))
	case "raw":
		fields = append(fields,
			gui.Input(gui.InputCfg{
				ID:   "body-ct",
				Text: body.ContentType,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					a := appCtx(ctx)
					a.DraftHTTP.Body.ContentType = s
					a.Dirty = true
				},
			}),
			gui.Input(gui.InputCfg{
				ID:     "body-raw",
				Mode:   gui.InputMultiline,
				Text:   body.Raw,
				Sizing: gui.FillFill,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					a := appCtx(ctx)
					a.DraftHTTP.Body.Kind = "raw"
					a.DraftHTTP.Body.Raw = s
					a.Dirty = true
				},
			}),
		)
	case "form":
		fields = append(fields, kvEditor("body-form", "form", body.Form, func(a *App, rows []httpapi.KV) {
			a.DraftHTTP.Body.Kind = "form"
			a.DraftHTTP.Body.Form = rows
		}))
	case "multipart":
		fields = append(fields, kvEditor("body-mp", "multipart (file:path)", body.Multipart, func(a *App, rows []httpapi.KV) {
			a.DraftHTTP.Body.Kind = "multipart"
			a.DraftHTTP.Body.Multipart = rows
		}))
	case "binary":
		fields = append(fields, gui.Input(gui.InputCfg{
			ID:   "body-bin",
			Text: body.BinaryPath,
			OnTextChanged: func(s string, ctx gui.EventCtx) {
				a := appCtx(ctx)
				a.DraftHTTP.Body.Kind = "binary"
				a.DraftHTTP.Body.BinaryPath = s
				a.Dirty = true
			},
		}))
	}
	return gui.Column(gui.ContainerCfg{Sizing: gui.FillFill, Spacing: gui.SpacingSmall, Content: fields})
}

func authEditor(app *App) gui.View {
	auth := app.DraftHTTP.Auth
	kind := "none"
	if auth != nil && auth.Kind != "" {
		kind = auth.Kind
	}
	fields := []gui.View{
		gui.Select(gui.SelectCfg{
			ID:       "auth-kind",
			Items:    []string{"none", "bearer", "basic", "apikey"},
			Selected: []string{kind},
			OnSelect: func(vals []string, ctx gui.EventCtx) {
				if len(vals) == 0 {
					return
				}
				a := appCtx(ctx)
				if vals[0] == "none" {
					a.DraftHTTP.Auth = nil
				} else {
					cur := a.DraftHTTP.Auth
					if cur == nil {
						cur = &httpapi.Auth{}
					}
					cur.Kind = vals[0]
					a.DraftHTTP.Auth = cur
				}
				a.Dirty = true
			},
		}),
	}
	if auth != nil {
		switch auth.Kind {
		case "bearer":
			fields = append(fields, gui.Input(gui.InputCfg{
				ID:   "auth-token",
				Text: auth.Token,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					a := appCtx(ctx)
					ensureAuth(a).Token = s
					a.Dirty = true
				},
			}))
		case "basic":
			fields = append(fields,
				gui.Input(gui.InputCfg{
					ID:   "auth-user",
					Text: auth.User,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						ensureAuth(a).User = s
						a.Dirty = true
					},
				}),
				gui.Input(gui.InputCfg{
					ID:   "auth-pass",
					Text: auth.Pass,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						ensureAuth(a).Pass = s
						a.Dirty = true
					},
				}),
			)
		case "apikey":
			fields = append(fields,
				gui.Input(gui.InputCfg{
					ID:   "auth-key",
					Text: auth.Key,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						ensureAuth(a).Key = s
						a.Dirty = true
					},
				}),
				gui.Input(gui.InputCfg{
					ID:   "auth-val",
					Text: auth.Value,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						ensureAuth(a).Value = s
						a.Dirty = true
					},
				}),
				gui.Select(gui.SelectCfg{
					ID:       "auth-in",
					Items:    []string{"header", "query"},
					Selected: selectedOrEmpty(authIn(auth.In)),
					OnSelect: func(vals []string, ctx gui.EventCtx) {
						if len(vals) == 0 {
							return
						}
						a := appCtx(ctx)
						ensureAuth(a).In = vals[0]
						a.Dirty = true
					},
				}),
			)
		}
	}
	return gui.Column(gui.ContainerCfg{Spacing: gui.SpacingSmall, Content: fields})
}

func optionsEditor(app *App) gui.View {
	cat := app.T()
	spec := app.DraftHTTP
	return gui.Column(gui.ContainerCfg{
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			gui.Input(gui.InputCfg{
				ID:    "opt-timeout",
				Label: cat.Timeout,
				Text:  strconv.Itoa(spec.TimeoutSec),
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					n, err := strconv.Atoi(strings.TrimSpace(s))
					if err != nil {
						return
					}
					a := appCtx(ctx)
					a.DraftHTTP.TimeoutSec = n
					a.Dirty = true
				},
			}),
			gui.LabeledToggle("opt-redirect", cat.FollowRedirects, spec.FollowRedirects, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				a.DraftHTTP.FollowRedirects = !a.DraftHTTP.FollowRedirects
				a.Dirty = true
			}),
			gui.LabeledToggle("opt-tls", cat.SkipTLS, spec.InsecureSkipVerify, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				a.DraftHTTP.InsecureSkipVerify = !a.DraftHTTP.InsecureSkipVerify
				a.Dirty = true
			}),
		},
	})
}

func sendCurrent(w *gui.Window) {
	a := appOf(w)
	if a.SelectedNodeID == 0 {
		return
	}
	n, err := a.Svc.GetNode(context.Background(), a.SelectedNodeID)
	if err != nil {
		a.fail(err)
		return
	}
	if n.Kind != store.NodeHTTP {
		return
	}
	raw, err := httpapi.EncodeSpec(a.DraftHTTP)
	if err != nil {
		a.fail(err)
		return
	}
	n.HTTP = raw

	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.sending = true
	a.Status = a.T().Sending
	a.mu.Unlock()

	go func() {
		res, err := a.Svc.SendHTTP(ctx, n, a.SelectedEnvID)
		a.mu.Lock()
		defer a.mu.Unlock()
		a.sending = false
		a.cancel = nil
		if err != nil {
			a.Status = err.Error()
			a.RespText = err.Error()
			return
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%s  %s\n", res.Status, res.Duration)
		if res.Truncated {
			b.WriteString("[truncated]\n")
		}
		b.Write(res.Body)
		a.RespText = b.String()
		a.Status = res.Status
	}()
}

func cancelSend(w *gui.Window) {
	a := appOf(w)
	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	a.sending = false
	a.Status = a.T().Cancel
	a.mu.Unlock()
}

func ensureAuth(a *App) *httpapi.Auth {
	if a.DraftHTTP.Auth == nil {
		a.DraftHTTP.Auth = &httpapi.Auth{}
	}
	return a.DraftHTTP.Auth
}

func authIn(in string) string {
	if in == "" {
		return "header"
	}
	return in
}
