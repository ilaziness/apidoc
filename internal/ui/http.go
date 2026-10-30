package ui

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-gui-org/go-gui/gui"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

func httpWorkspace(app *App) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Padding: gui.PaddingSmall,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			editorNameRow(app),
			gui.TabControl(gui.TabControlCfg{
				ID:       "http-work-tabs",
				Sizing:   gui.FillFill,
				Selected: app.WorkTab,
				OnSelect: func(id string, ctx gui.EventCtx) {
					appCtx(ctx).WorkTab = id
				},
				Items: []gui.TabItemCfg{
					gui.NewTabItem(workPreview, cat.Preview, []gui.View{httpPreview(app)}),
					gui.NewTabItem(workEdit, cat.EditDoc, []gui.View{httpEdit(app)}),
					gui.NewTabItem(workDebug, cat.Debug, []gui.View{httpDebug(app)}),
				},
			}),
		},
	})
}

func httpPreview(app *App) gui.View {
	cat := app.T()
	spec := app.DraftHTTP
	desc := spec.Description
	if desc == "" {
		desc = "—"
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Padding: gui.PaddingSmall,
		Content: []gui.View{
			gui.Label(cat.Method+": "+spec.Method, gui.CurrentTheme().TextStyleTitleSmall),
			gui.Input(gui.InputCfg{
				ID:       "http-preview-url",
				Label:    cat.URL,
				Text:     spec.URL,
				ReadOnly: true,
				Sizing:   gui.FillFit,
			}),
			gui.Input(gui.InputCfg{
				ID:       "http-preview-desc",
				Label:    cat.Description,
				Mode:     gui.InputMultiline,
				Text:     desc,
				ReadOnly: true,
				Sizing:   gui.FillFill,
			}),
		},
	})
}

func httpEdit(app *App) gui.View {
	cat := app.T()
	spec := app.DraftHTTP
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SpacingSmall,
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{
					gui.Select(gui.SelectCfg{
						ID:       "http-edit-method",
						Label:    cat.Method,
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
					gui.Input(gui.InputCfg{
						ID:     "http-edit-url",
						Label:  cat.URL,
						Text:   spec.URL,
						Sizing: gui.FillFit,
						OnTextChanged: func(s string, ctx gui.EventCtx) {
							a := appCtx(ctx)
							a.DraftHTTP.URL = s
							a.Dirty = true
						},
					}),
				},
			}),
			gui.Input(gui.InputCfg{
				ID:     "http-edit-desc",
				Label:  cat.Description,
				Mode:   gui.InputMultiline,
				Text:   spec.Description,
				Sizing: gui.FillFit,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					a := appCtx(ctx)
					a.DraftHTTP.Description = s
					a.Dirty = true
				},
			}),
			httpTabs(app, "edit"),
		},
	})
}

func httpDebug(app *App) gui.View {
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
				ID:       "http-debug-method",
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
				ID:     "http-debug-url",
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
					Content: []gui.View{httpTabs(app, "debug")},
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

func httpTabs(app *App, prefix string) gui.View {
	cat := app.T()
	return gui.TabControl(gui.TabControlCfg{
		ID:       prefix + "-http-tabs",
		Sizing:   gui.FillFill,
		Selected: app.EditorTab,
		OnSelect: func(id string, ctx gui.EventCtx) {
			appCtx(ctx).EditorTab = id
		},
		Items: []gui.TabItemCfg{
			gui.NewTabItem(tabParams, cat.Params, []gui.View{kvPairEditors(app, prefix)}),
			gui.NewTabItem(tabHeaders, cat.Headers, []gui.View{
				kvEditor(prefix+"-http-headers", cat.Headers, app.DraftHTTP.Headers, func(a *App, rows []httpapi.KV) {
					a.DraftHTTP.Headers = rows
				}),
			}),
			gui.NewTabItem(tabBody, cat.Body, []gui.View{bodyEditor(app, prefix)}),
			gui.NewTabItem(tabAuth, cat.Auth, []gui.View{authEditor(app, prefix)}),
			gui.NewTabItem(tabOptions, cat.Options, []gui.View{optionsEditor(app, prefix)}),
		},
	})
}

func kvPairEditors(app *App, prefix string) gui.View {
	cat := app.T()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			kvEditor(prefix+"-http-query", cat.Query, app.DraftHTTP.Query, func(a *App, rows []httpapi.KV) {
				a.DraftHTTP.Query = rows
			}),
			kvEditor(prefix+"-http-path", cat.PathParams, app.DraftHTTP.PathParams, func(a *App, rows []httpapi.KV) {
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

func bodyEditor(app *App, prefix string) gui.View {
	kind := app.DraftHTTP.Body.Kind
	if kind == "" {
		kind = "none"
	}
	body := app.DraftHTTP.Body
	fields := []gui.View{
		gui.Select(gui.SelectCfg{
			ID:       prefix + "-body-kind",
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
			ID:     prefix + "-body-json",
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
				ID:   prefix + "-body-ct",
				Text: body.ContentType,
				OnTextChanged: func(s string, ctx gui.EventCtx) {
					a := appCtx(ctx)
					a.DraftHTTP.Body.ContentType = s
					a.Dirty = true
				},
			}),
			gui.Input(gui.InputCfg{
				ID:     prefix + "-body-raw",
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
		fields = append(fields, kvEditor(prefix+"-body-form", "form", body.Form, func(a *App, rows []httpapi.KV) {
			a.DraftHTTP.Body.Kind = "form"
			a.DraftHTTP.Body.Form = rows
		}))
	case "multipart":
		fields = append(fields, kvEditor(prefix+"-body-mp", "multipart (file:path)", body.Multipart, func(a *App, rows []httpapi.KV) {
			a.DraftHTTP.Body.Kind = "multipart"
			a.DraftHTTP.Body.Multipart = rows
		}))
	case "binary":
		fields = append(fields, gui.Input(gui.InputCfg{
			ID:   prefix + "-body-bin",
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

func authEditor(app *App, prefix string) gui.View {
	auth := app.DraftHTTP.Auth
	kind := "none"
	if auth != nil && auth.Kind != "" {
		kind = auth.Kind
	}
	fields := []gui.View{
		gui.Select(gui.SelectCfg{
			ID:       prefix + "-auth-kind",
			Items:    []string{"none", "bearer", "basic", "apikey"},
			Selected: []string{kind},
			OnSelect: func(vals []string, ctx gui.EventCtx) {
				if len(vals) == 0 {
					return
				}
				a := appCtx(ctx)
				if vals[0] == "none" {
					a.DraftHTTP.Auth = &httpapi.Auth{Kind: "none"}
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
				ID:   prefix + "-auth-token",
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
					ID:   prefix + "-auth-user",
					Text: auth.User,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						ensureAuth(a).User = s
						a.Dirty = true
					},
				}),
				gui.Input(gui.InputCfg{
					ID:   prefix + "-auth-pass",
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
					ID:   prefix + "-auth-key",
					Text: auth.Key,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						ensureAuth(a).Key = s
						a.Dirty = true
					},
				}),
				gui.Input(gui.InputCfg{
					ID:   prefix + "-auth-val",
					Text: auth.Value,
					OnTextChanged: func(s string, ctx gui.EventCtx) {
						a := appCtx(ctx)
						ensureAuth(a).Value = s
						a.Dirty = true
					},
				}),
				gui.Select(gui.SelectCfg{
					ID:       prefix + "-auth-in",
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

func optionsEditor(app *App, prefix string) gui.View {
	cat := app.T()
	spec := app.DraftHTTP
	return gui.Column(gui.ContainerCfg{
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			gui.Input(gui.InputCfg{
				ID:    prefix + "-opt-timeout",
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
			gui.LabeledToggle(prefix+"-opt-redirect", cat.FollowRedirects, spec.FollowRedirects, func(ctx gui.EventCtx) {
				a := appCtx(ctx)
				a.DraftHTTP.FollowRedirects = !a.DraftHTTP.FollowRedirects
				a.Dirty = true
			}),
			gui.LabeledToggle(prefix+"-opt-tls", cat.SkipTLS, spec.InsecureSkipVerify, func(ctx gui.EventCtx) {
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
	nodeID := a.SelectedNodeID
	envID := a.SelectedEnvID

	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.sendGen++
	gen := a.sendGen
	a.sending = true
	a.Status = a.T().Sending
	a.mu.Unlock()

	go func() {
		res, err := a.Svc.SendHTTP(ctx, n, envID)
		w.QueueCommand(func(win *gui.Window) {
			app := appOf(win)
			app.mu.Lock()
			mine := app.sendGen == gen
			if mine {
				app.sending = false
				app.cancel = nil
			}
			app.mu.Unlock()
			if !mine || app.SelectedNodeID != nodeID {
				return
			}
			if err != nil {
				app.Status = err.Error()
				app.RespText = err.Error()
				return
			}
			app.RespText = formatHTTPResponse(res)
			app.Status = res.Status
		})
	}()
}

func formatHTTPResponse(res *httpapi.Response) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n", res.Status, res.Duration)
	keys := make([]string, 0, len(res.Headers))
	for k := range res.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range res.Headers[k] {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	if res.Truncated {
		b.WriteString("[truncated]\n")
	}
	b.WriteByte('\n')
	b.Write(res.Body)
	return b.String()
}

func cancelSend(w *gui.Window) {
	a := appOf(w)
	a.stopSend()
	a.Status = a.T().Cancel
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
