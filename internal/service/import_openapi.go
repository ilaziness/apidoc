package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ilaziness/apidoc/internal/config"
	"github.com/ilaziness/apidoc/internal/openapi"
	"github.com/ilaziness/apidoc/internal/protocol"
	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
	"github.com/ilaziness/apidoc/internal/store"
)

// ImportOpenAPIOptions controls import target.
type ImportOpenAPIOptions struct {
	Source           openapi.Source
	IntoCollectionID int64 // 0 = create new collection
	UntaggedLabel    string
}

// ImportOpenAPI imports operations into a new or existing collection (overwrite on name clash).
func (s *Service) ImportOpenAPI(ctx context.Context, opt ImportOpenAPIOptions) (*store.Collection, error) {
	doc, err := openapi.Import(ctx, opt.Source)
	if err != nil {
		return nil, err
	}
	label := opt.UntaggedLabel
	if label == "" {
		label = UntaggedFolderName(s.cfg.Language)
	}

	var col *store.Collection
	if opt.IntoCollectionID == 0 {
		name := doc.Title
		if name == "" {
			name = "Imported API"
		}
		col, err = s.store.CreateCollection(ctx, name)
		if err != nil {
			return nil, err
		}
	} else {
		col, err = s.store.GetCollection(ctx, opt.IntoCollectionID, false)
		if err != nil {
			return nil, err
		}
	}

	folderIDs := map[string]int64{}
	for _, op := range doc.Operations {
		tag := label
		if len(op.Tags) > 0 && op.Tags[0] != "" {
			tag = op.Tags[0]
		}
		fid, ok := folderIDs[tag]
		if !ok {
			existing, err := s.store.FindNodeByName(ctx, col.ID, nil, store.NodeFolder, tag)
			if err == nil {
				fid = existing.ID
			} else if errors.Is(err, sql.ErrNoRows) {
				folder := &store.Node{CollectionID: col.ID, Kind: store.NodeFolder, Name: tag}
				if err := s.store.CreateNode(ctx, folder); err != nil {
					return nil, err
				}
				fid = folder.ID
			} else {
				return nil, err
			}
			folderIDs[tag] = fid
		}
		parent := fid
		spec := httpapi.DefaultSpec()
		spec.Method = op.Method
		spec.URL = op.Path
		spec.Description = op.Description
		raw, err := httpapi.EncodeSpec(spec)
		if err != nil {
			return nil, err
		}
		existing, err := s.store.FindNodeByName(ctx, col.ID, &parent, store.NodeHTTP, op.Name)
		if err == nil {
			existing.HTTP = raw
			existing.Protocol = string(protocol.HTTP)
			if err := s.store.UpdateNode(ctx, existing); err != nil {
				return nil, err
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		n := &store.Node{
			CollectionID: col.ID,
			ParentID:     &parent,
			Kind:         store.NodeHTTP,
			Name:         op.Name,
			Protocol:     string(protocol.HTTP),
			HTTP:         raw,
		}
		if err := s.store.CreateNode(ctx, n); err != nil {
			return nil, fmt.Errorf("import op %s: %w", op.Name, err)
		}
	}
	return col, nil
}

// UntaggedFolderName returns the localized untagged folder label.
func UntaggedFolderName(lang string) string {
	switch lang {
	case config.LangZHHans:
		return "未分组"
	case config.LangZHHant:
		return "未分組"
	default:
		return "Untagged"
	}
}
