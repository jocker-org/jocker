package parser

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/containerd/platforms"
	"github.com/google/go-jsonnet"
	"github.com/jocker-org/jocker/dockerignore"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/pkg/errors"
)

func readFile(ctx context.Context, c client.Client, filename string) (content []byte, err error) {
	src := llb.Local("context",
		llb.IncludePatterns([]string{
			filename,
			"**/" + filename,
			"*/",
		}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint("Jockerfile"),
		llb.WithCustomName("load " + filename),

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

func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	opts := c.BuildOpts().Opts
	filename := opts["filename"]
	if filename == "" {
		filename = "Jockerfile"
	}

	buildargs := make(map[string]string)
	for k, v := range opts {
		if strings.HasPrefix(k, "build-arg:") {
			buildargs[k[10:]] = v
		}
	}

	jopts, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal opts: %w", err)
	}

	jbuildargs, err := json.Marshal(buildargs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal buildargs: %w", err)
	}

	vm := jsonnet.MakeVM()
	vm.ExtCode("buildArgs", string(jbuildargs))
	vm.ExtCode("filename", fmt.Sprintf("%q",filename))
	vm.ExtCode("opts", fmt.Sprintf("%q",jopts))
	vm.Importer(NewChainedImporter(NewContextImporter(ctx, c), []string{"/lib/"}))

	dtJockerfile, err := readFile(ctx, c, filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Jockerfile")
	}

	// jsonStr, err := vm.EvaluateFile(filename)
	jsonStr, err := vm.EvaluateAnonymousSnippet(filename, string(dtJockerfile))
	if err != nil {
		return nil, err
	}

	j, err := ParseJockerfile(jsonStr)
	if err != nil {
		return nil, err
	}

	if len(j.Excludes) == 0 {
		content, err := readFile(ctx, c, ".dockerignore")
		if err != nil {
			j.Excludes, _ = dockerignore.Parse(bytes.NewReader(content))
		}
	}

	state := j.ToLLB(buildargs["debug"], ctx, c)

	dt, err := state.Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		return nil, err
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: dt.ToPB(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to resolve jockerfile: %w", err)
	}

	ref, err := res.SingleRef()
	if err != nil {
		return nil, err
	}

	p := platforms.DefaultSpec()
	if userplatform, ok := opts["platform"]; ok {
		if p, err = platforms.Parse(userplatform); err != nil {
			return nil, fmt.Errorf("failed to parse platform %s: %w",
				userplatform, err)
		}
	}

	j.Image.Platform = p

	config, err := json.Marshal(j.Image)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal image config")
	}
	res.SetRef(ref)
	res.AddMeta(exptypes.ExporterImageConfigKey, config)

	return res, nil
}
