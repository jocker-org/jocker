package parser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containerd/platforms"
	"github.com/google/go-jsonnet"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func ReadFile(ctx context.Context, c client.Client, filename string) (string, error) {
	src := llb.Local("context",
		llb.IncludePatterns([]string{filename}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint("Jockerfile"),
	)

	def, err := src.Marshal(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal local source: %w", err)
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
	})
	if err != nil {
		return "", err
	}

	ref, err := res.SingleRef()
	if err != nil {
		return "", err
	}

	content, err := ref.ReadFile(ctx, client.ReadRequest{
		Filename: filename,
	})
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	opts := c.BuildOpts().Opts
	filename := opts["filename"]
	if filename == "" {
		filename = "Jockerfile"
	}

	jockerfile, err := ReadFile(ctx, c, filename)
	if err != nil {
		return nil, err
	}

	vm := jsonnet.MakeVM()
	jsonStr, err := vm.EvaluateAnonymousSnippet(filename, jockerfile)
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
