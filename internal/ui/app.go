package ui

import (
	"context"
	"sync"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/i18n"
	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/service"
	"github.com/ilaziness/apidoc/internal/store"
)

type panel int

const (
	panelEditor panel = iota
	panelSettings
	panelTrash
	panelImport
	panelEnv
)

const (
	tabParams  = "params"
	tabHeaders = "headers"
	tabBody    = "body"
	tabAuth    = "auth"
	tabOptions = "options"
)

// App is the go-gui window state: one session, one collection at a time.
type App struct {
	Svc *service.Service

	Collections []store.Collection
	Nodes       []store.Node
	Envs        []store.Environment

	SelectedCollectionID int64
	SelectedNodeID       int64
	SelectedEnvID        int64

	DraftName     string
	DraftMarkdown string
	DraftHTTP     httpapi.Spec
	Dirty         bool

	EditorTab string
	SideRatio float32
	ReqRatio  float32

	Status   string
	RespText string
	Panel    panel

	SetLanguage  string
	SetTheme     string
	SetProxyMode string
	SetProxyURL  string
	SetDataDir   string

	ImportPath string
	ImportURL  string

	DraftEnvName string
	DraftEnvVars string

	sending bool
	mu      sync.Mutex
	cancel  context.CancelFunc
}

func newApp(svc *service.Service) *App {
	return &App{
		Svc:       svc,
		Panel:     panelEditor,
		EditorTab: tabParams,
		SideRatio: 0.22,
		ReqRatio:  0.55,
	}
}

func (a *App) T() i18n.Catalog {
	return i18n.T(a.Svc.Config().Language)
}

func (a *App) loadSettingsDraft(cfg *config.Config) {
	a.SetLanguage = cfg.Language
	a.SetTheme = cfg.Theme
	a.SetProxyMode = cfg.Proxy.Mode
	a.SetProxyURL = cfg.Proxy.Manual
	a.SetDataDir = cfg.DataDir
}

func (a *App) selectedKind() string {
	for _, n := range a.Nodes {
		if n.ID == a.SelectedNodeID {
			return n.Kind
		}
	}
	return ""
}

func (a *App) selectedCollectionName() string {
	for _, c := range a.Collections {
		if c.ID == a.SelectedCollectionID {
			return c.Name
		}
	}
	return ""
}

func (a *App) insertParent() *int64 {
	for _, n := range a.Nodes {
		if n.ID != a.SelectedNodeID {
			continue
		}
		if n.Kind == store.NodeFolder {
			id := n.ID
			return &id
		}
		if n.ParentID != nil {
			id := *n.ParentID
			return &id
		}
		return nil
	}
	return nil
}

func (a *App) refreshCollections(ctx context.Context) {
	list, err := a.Svc.ListCollections(ctx)
	if err != nil {
		a.Status = err.Error()
		return
	}
	a.Collections = list
}

func (a *App) refreshTree(ctx context.Context) {
	if a.SelectedCollectionID == 0 {
		a.Nodes = nil
		a.Envs = nil
		return
	}
	nodes, err := a.Svc.ListNodes(ctx, a.SelectedCollectionID)
	if err != nil {
		a.Status = err.Error()
		return
	}
	a.Nodes = nodes
	envs, err := a.Svc.ListEnvironments(ctx, a.SelectedCollectionID)
	if err != nil {
		a.Status = err.Error()
		return
	}
	a.Envs = envs
}

func (a *App) syncEnvDraft() {
	a.DraftEnvName = ""
	a.DraftEnvVars = ""
	for _, e := range a.Envs {
		if e.ID == a.SelectedEnvID {
			a.DraftEnvName = e.Name
			a.DraftEnvVars = encodeMap(service.ParseEnvVars(e.Variables))
			return
		}
	}
}

func (a *App) selectCollection(ctx context.Context, id int64) {
	preferredEnv := a.SelectedEnvID
	if preferredEnv == 0 {
		preferredEnv = a.Svc.Config().LastEnvironmentID
	}

	a.SelectedCollectionID = id
	a.SelectedNodeID = 0
	a.Dirty = false
	a.DraftName = ""
	a.DraftMarkdown = ""
	a.DraftHTTP = httpapi.DefaultSpec()
	a.RespText = ""
	a.Panel = panelEditor
	a.refreshTree(ctx)

	a.SelectedEnvID = 0
	for _, e := range a.Envs {
		if e.ID == preferredEnv {
			a.SelectedEnvID = e.ID
			break
		}
	}
	if a.SelectedEnvID == 0 && len(a.Envs) > 0 {
		a.SelectedEnvID = a.Envs[0].ID
	}
	a.syncEnvDraft()

	cfg := a.Svc.Config()
	cfg.LastCollectionID = id
	cfg.LastEnvironmentID = a.SelectedEnvID
	_ = a.Svc.SaveConfig()
}

func (a *App) loadNode(ctx context.Context, id int64) {
	n, err := a.Svc.GetNode(ctx, id)
	if err != nil {
		a.Status = err.Error()
		return
	}
	a.SelectedNodeID = id
	a.DraftName = n.Name
	a.Dirty = false
	a.RespText = ""
	a.DraftMarkdown = ""
	a.DraftHTTP = httpapi.DefaultSpec()
	a.Panel = panelEditor
	switch n.Kind {
	case store.NodeMarkdown:
		a.DraftMarkdown = n.Markdown
	case store.NodeHTTP:
		spec, err := httpapi.DecodeSpec(n.HTTP)
		if err != nil {
			a.Status = err.Error()
			return
		}
		a.DraftHTTP = spec
	}
}

func (a *App) discardDraft(ctx context.Context) {
	if a.SelectedNodeID == 0 {
		a.Dirty = false
		return
	}
	a.loadNode(ctx, a.SelectedNodeID)
	a.Status = ""
}

func (a *App) fail(err error) {
	a.Status = err.Error()
}

func (a *App) isSending() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sending
}
