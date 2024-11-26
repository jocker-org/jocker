package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	jsonStr, err := EvaluateJsonnetFile("Jockerfile")
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

	imgConfig := specs.ImageConfig{
		Cmd: []string {"ls"},
	}

	img := &specs.Image{
		Platform: specs.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
		Config: imgConfig,
	}

	config, err := json.Marshal(img)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image config: %w", err)
	}
	k := platforms.Format(platforms.DefaultSpec())

	res.AddMeta(fmt.Sprintf("%s/%s", exptypes.ExporterImageConfigKey, k), config)
	res.SetRef(ref)

	return res, nil
}
