package parser

import (
	"context"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/moby/buildkit/frontend/gateway/client"
)

type ContexImporter struct {
	ctx   context.Context
	cache map[string]jsonnet.Contents
	c     client.Client
}

func NewContextImporter(ctx context.Context, c client.Client) *ContexImporter {
	return &ContexImporter{
		ctx:   ctx,
		cache: make(map[string]jsonnet.Contents),
		c:     c,
	}
}

func (importer *ContexImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundHere string, err error) {
	dir, _ := filepath.Split(importedFrom)

	absPath := importedPath
	if !filepath.IsAbs(importedPath) {
		absPath = filepath.Join(dir, importedPath)
	}

	if entry, ok := importer.cache[absPath]; ok {
		return entry, absPath, nil
	}

	content, err := readFile(importer.ctx, importer.c, absPath)
	if err != nil {
		// TODO: distinguish between file not found and other
		// failures?
		return
	}
	entry := jsonnet.MakeContentsRaw(content)
	importer.cache[absPath] = entry
	return entry, absPath, nil
}
