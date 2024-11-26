package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/containerd/platforms"
	"github.com/google/go-jsonnet"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func GetJockerfile(ctx context.Context, c client.Client) (string, error) {
	opts := c.BuildOpts().Opts
	filename := opts["filename"]
	if filename == "" {
		filename = "Jockerfile"
	}

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

	jockerfile, err := ref.ReadFile(ctx, client.ReadRequest{
		Filename: filename,
	})
	if err != nil {
		return "", err
	}
	return string(jockerfile), nil
}

func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	jockerfile, err := GetJockerfile(ctx, c)
	if err != nil {
		log.Fatal(err)
	}

	vm := jsonnet.MakeVM()
	jsonStr, err := vm.EvaluateAnonymousSnippet("jockerfile", jockerfile)
	if err != nil {
		log.Fatal(err)
	}

	j, err := ParseJockerfile(jsonStr)
	if err != nil {
		log.Fatal(err)
	}

	state := j.ToLLB()
	dt, err := state.Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		log.Fatal(err)
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
			Cmd: []string {"/bin/ls"},
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
