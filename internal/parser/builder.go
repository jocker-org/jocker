package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/containerd/platforms"
	"github.com/google/go-jsonnet"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func readFile(ctx context.Context, c client.Client, filename string) (content []byte, err error) {
	src := llb.Local("context",
		llb.IncludePatterns([]string{filename}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint("Jockerfile"),
	)

	def, err := src.Marshal(ctx)
	if err != nil {
		return
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
	})
	if err != nil {
		return
	}

	ref, err := res.SingleRef()
	if err != nil {
		return
	}

	return ref.ReadFile(ctx, client.ReadRequest{
		Filename: filename,
	})
}

type ContexImporter struct {
	ctx context.Context
	cache map[string]jsonnet.Contents
	c client.Client
}

func NewContextImporter(ctx context.Context, c client.Client) *ContexImporter {
	return &ContexImporter{
		ctx: ctx,
		cache: make(map[string]jsonnet.Contents),
		c: c,
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

func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	opts := c.BuildOpts().Opts
	filename := opts["filename"]
	if filename == "" {
		filename = "Jockerfile"
	}

	importer := NewContextImporter(ctx, c)
	content, _, err := importer.Import("", filename)
	if err != nil {
		return nil, err
	}

	vm := jsonnet.MakeVM()
	vm.Importer(importer)
	jsonStr, err := vm.EvaluateAnonymousSnippet(filename, content.String())
	if err != nil {
		return nil, err
	}

	j, err := ParseJockerfile(jsonStr)
	if err != nil {
		return nil, err
	}

	state := j.ToLLB()
	dt, err := state.Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		return nil, err
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: dt.ToPB(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to resolve dockerfile: %w", err)
	}

	ref, err := res.SingleRef()
	if err != nil {
		return nil, err
	}

	p := platforms.DefaultSpec()
	img := &specs.Image{
		Platform: p,
		Config: specs.ImageConfig{
			Cmd: j.Cmd,
		},
	}

	config, err := json.Marshal(img)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image config: %w", err)
	}
	res.AddMeta(exptypes.ExporterImageConfigKey, config)
	res.SetRef(ref)

	return res, nil
}
