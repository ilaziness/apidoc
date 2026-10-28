// Package service holds application use cases.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/protocol"
	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

const trashRetention = 30 * 24 * time.Hour

// Deps are collaborators constructed by main.
type Deps struct {
	Store  *store.DB
	Config *config.Config
	HTTP   *httpapi.Handler
}

// Service is the application facade used by the UI.
type Service struct {
	store *store.DB
	cfg   *config.Config
	http  *httpapi.Handler
}

// New builds a Service from deps.
func New(deps Deps) *Service {
	return &Service{store: deps.Store, cfg: deps.Config, http: deps.HTTP}
}

// Config returns the live app config pointer.
func (s *Service) Config() *config.Config { return s.cfg }

// SaveConfig persists settings.
func (s *Service) SaveConfig() error { return config.Save(s.cfg) }

// PurgeTrash removes expired soft-deleted rows.
func (s *Service) PurgeTrash(ctx context.Context) error {
	return s.store.PurgeExpiredTrash(ctx, trashRetention)
}

// --- collections ---

func (s *Service) ListCollections(ctx context.Context) ([]store.Collection, error) {
	return s.store.ListCollections(ctx)
}

func (s *Service) CreateCollection(ctx context.Context, name string) (*store.Collection, error) {
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	return s.store.CreateCollection(ctx, name)
}

func (s *Service) RenameCollection(ctx context.Context, id int64, name string) error {
	return s.store.RenameCollection(ctx, id, name)
}

func (s *Service) DeleteCollection(ctx context.Context, id int64) error {
	return s.store.SoftDeleteCollection(ctx, id)
}

func (s *Service) RestoreCollection(ctx context.Context, id int64) error {
	return s.store.RestoreCollection(ctx, id)
}

func (s *Service) ListTrashCollections(ctx context.Context) ([]store.Collection, error) {
	return s.store.ListTrashedCollections(ctx)
}

func (s *Service) EmptyTrash(ctx context.Context) error {
	return s.store.EmptyTrash(ctx)
}

// --- nodes ---

func (s *Service) ListNodes(ctx context.Context, collectionID int64) ([]store.Node, error) {
	return s.store.ListNodes(ctx, collectionID)
}

func (s *Service) GetNode(ctx context.Context, id int64) (*store.Node, error) {
	return s.store.GetNode(ctx, id)
}

func (s *Service) CreateFolder(ctx context.Context, collectionID int64, parentID *int64, name string) (*store.Node, error) {
	n := &store.Node{CollectionID: collectionID, ParentID: parentID, Kind: store.NodeFolder, Name: name}
	if err := s.store.CreateNode(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) CreateMarkdown(ctx context.Context, collectionID int64, parentID *int64, name, body string) (*store.Node, error) {
	n := &store.Node{CollectionID: collectionID, ParentID: parentID, Kind: store.NodeMarkdown, Name: name, Markdown: body}
	if err := s.store.CreateNode(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) CreateHTTP(ctx context.Context, collectionID int64, parentID *int64, name string, spec httpapi.Spec) (*store.Node, error) {
	raw, err := httpapi.EncodeSpec(spec)
	if err != nil {
		return nil, err
	}
	n := &store.Node{
		CollectionID: collectionID,
		ParentID:     parentID,
		Kind:         store.NodeHTTP,
		Name:         name,
		Protocol:     string(protocol.HTTP),
		HTTP:         raw,
	}
	if err := s.store.CreateNode(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) SaveNode(ctx context.Context, n *store.Node) error {
	return s.store.UpdateNode(ctx, n)
}

func (s *Service) DeleteNode(ctx context.Context, id int64) error {
	return s.store.SoftDeleteNode(ctx, id)
}

func (s *Service) ListTrashNodes(ctx context.Context) ([]store.Node, error) {
	return s.store.ListTrashedNodes(ctx)
}

func (s *Service) RestoreNode(ctx context.Context, id int64) error {
	return s.store.RestoreNode(ctx, id)
}

// --- environments ---

func (s *Service) ListEnvironments(ctx context.Context, collectionID int64) ([]store.Environment, error) {
	return s.store.ListEnvironments(ctx, collectionID)
}

func (s *Service) GetEnvironment(ctx context.Context, id int64) (*store.Environment, error) {
	return s.store.GetEnvironment(ctx, id)
}

func (s *Service) CreateEnvironment(ctx context.Context, collectionID int64, name string, vars map[string]string) (*store.Environment, error) {
	raw, err := json.Marshal(vars)
	if err != nil {
		return nil, err
	}
	return s.store.CreateEnvironment(ctx, collectionID, name, string(raw))
}

func (s *Service) UpdateEnvironment(ctx context.Context, e *store.Environment, vars map[string]string) error {
	raw, err := json.Marshal(vars)
	if err != nil {
		return err
	}
	e.Variables = string(raw)
	return s.store.UpdateEnvironment(ctx, e)
}

func (s *Service) DeleteEnvironment(ctx context.Context, id int64) error {
	return s.store.SoftDeleteEnvironment(ctx, id)
}

func ParseEnvVars(raw string) map[string]string {
	out := map[string]string{}
	if raw == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	if out == nil {
		out = map[string]string{}
	}
	return out
}

// SendHTTP loads env + collection default auth and sends.
func (s *Service) SendHTTP(ctx context.Context, node *store.Node, envID int64) (*httpapi.Response, error) {
	if node.Kind != store.NodeHTTP {
		return nil, fmt.Errorf("not an http node")
	}
	spec, err := httpapi.DecodeSpec(node.HTTP)
	if err != nil {
		return nil, err
	}
	var envMap map[string]string
	if envID > 0 {
		e, err := s.store.GetEnvironment(ctx, envID)
		if err != nil {
			return nil, err
		}
		envMap = ParseEnvVars(e.Variables)
	}
	col, err := s.store.GetCollection(ctx, node.CollectionID, false)
	if err != nil {
		return nil, err
	}
	var colAuth *httpapi.Auth
	if col.DefaultAuth != "" {
		colAuth = &httpapi.Auth{}
		_ = json.Unmarshal([]byte(col.DefaultAuth), colAuth)
	}
	return s.http.Send(ctx, spec, httpapi.SendOptions{
		Proxy:          s.cfg.Proxy,
		Env:            envMap,
		CollectionAuth: colAuth,
	})
}
