package parser

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	"github.com/moby/buildkit/client/llb/sourceresolver"
	"github.com/moby/buildkit/frontend/gateway/client"

	gw "github.com/moby/buildkit/frontend/gateway/client"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
)

type BuildContext struct {
	stages  map[string]llb.State
	state   llb.State
	context llb.State
	debug   string
	ctx     context.Context
	image dockerspec.DockerOCIImage
}

type BuildStep interface {
	Evaluate(*BuildContext) llb.State
}

func parseKeyValue(env string) (string, string) {
	parts := strings.SplitN(env, "=", 2)
	v := ""
	if len(parts) > 1 {
		v = parts[1]
	}

	return parts[0], v
}

// normalizeImage returns the image with the docker.io/library prefixed if it's not
// from a registry. Otherwise we can't resolve the Image Metadata
func normalizeImage(image string) string {
	parts := strings.Split(image, "/")

	if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
		return image
	}

	if len(parts) == 1 {
		return "docker.io/library/" + image
	}

	return "docker.io/" + image
}

// debugLog conditionally logs a debug message and injects an echo command into the LLB state.
// it also adds a random uuid for invalidating docker cache
// https://github.com/docker/buildx/issues/2387
func debugLog(b *BuildContext, msg string) {
	by := make([]byte, 4)
	_, err := rand.Read(by)
	if err != nil {
		slog.Error("error while creating unique id", "error", err)
	}
	uuid := fmt.Sprintf("%x ", by)
	if b.debug == "all" {
		slog.Warn("Step: ", "buildctx", msg)
		b.state = b.state.Run(llb.Shlex("echo DEBUG " + uuid + msg)).Root()
	}
}

func (c *ArgStep) Evaluate(b *BuildContext) llb.State {
	b.state = b.state.AddEnv(c.Name, c.Value)
	return b.state
}

func (c *CopyStep) Evaluate(b *BuildContext) llb.State {
	st := b.context
	if c.From != "" {
		st = b.stages[c.From]
	}
	if c.Source == "" || c.Destination == "" {
		return b.state
	}
	// by default mimick the Dockerfile behaviour, copying the
	// content only, and not the directory itself
	copyInfo := llb.CopyInfo{
		CopyDirContentsOnly: true,
	}

	opts := []llb.CopyOption{
		&copyInfo,
	}

	b.state = b.state.File(llb.Copy(st, c.Source, c.Destination, opts...))
	return b.state
}

func (c *RunStep) Evaluate(b *BuildContext) llb.State {
	if c.Command == "" {
		return b.state
	}

	debugLog(b, c.Command)

	b.state = b.state.Run(shf(c.Command)).Root()
	return b.state
}

func (c *WorkdirStep) Evaluate(b *BuildContext) llb.State {
	debugLog(b, c.Path)
	b.state = b.state.Dir(c.Path)
	return b.state
}

func (c *UserStep) Evaluate(b *BuildContext) llb.State {
	b.state = b.state.With(llb.User(c.User))
	return b.state
}

func shf(cmd string, v ...interface{}) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", fmt.Sprintf(cmd, v...)})
}

func (stage *BuildStage) ToLLB(b *BuildContext, c client.Client) llb.State {
	if stage.From == "scratch" {
		b.state = llb.Scratch()
	} else {
		var img dockerspec.DockerOCIImage
		baseImg := normalizeImage(stage.From)
		if c == nil {
			metaresolver := imagemetaresolver.Default()
			_, _, dt, err := metaresolver.ResolveImageConfig(b.ctx, baseImg, sourceresolver.Opt{
				ImageOpt: &sourceresolver.ResolveImageOpt{
					ResolveMode: llb.ResolveModeDefault.String(),
				},
			})
			if err != nil {
				debugLog(b, "failed to resolve image")
				slog.Error("failed to resolve image", "FROM", err)
			}
			if err := json.Unmarshal(dt, &img); err != nil {
				debugLog(b, "failed to unmarshal image")
				slog.Error("failed to unmarshal image", "FROM", err)
			}
		} else {
			_, _, dt, err := gw.Client.ResolveImageConfig(c, b.ctx, baseImg, sourceresolver.Opt{
				ImageOpt: &sourceresolver.ResolveImageOpt{
					ResolveMode: llb.ResolveModeDefault.String(),
				},
			})
			if err != nil {
				debugLog(b, "failed to resolve image")
				slog.Error("failed to resolve image", "FROM", err)
			}
			if err := json.Unmarshal(dt, &img); err != nil {
				debugLog(b, "failed to unmarshal image")
				slog.Error("failed to unmarshal image", "FROM", err)
			}

		}
		b.state = llb.Image(stage.From)
		b.image = img

		// initialize metadata
		for _, env := range img.Config.Env {
			debugLog(b, env)
			k, v := parseKeyValue(env)
			b.state = b.state.AddEnv(k, v)
		}
	}

	for i := range *stage.Steps {
		slog.Info("Building steps", "stage", (*stage.Steps)[i])
		b.state = (*stage.Steps)[i].Evaluate(b)
	}

	return b.state
}

func (j *Jockerfile) ToLLB(debug string, ctx context.Context, c client.Client) llb.State {
	b := BuildContext{
		stages: make(map[string]llb.State),
		debug:  debug,
		ctx: ctx,
	}
	var state llb.State
	opts := []llb.LocalOption{
		llb.ExcludePatterns(j.Excludes),
	}

	b.context = llb.Local("context", opts...)

	for _, stage := range j.Stages {
		slog.Info("Building stage", "ctx", stage.Name)
		state = stage.ToLLB(&b, c)
		b.stages[stage.Name] = state
	}

	// after all stages align the imageConfig to export
	slog.Info("setting image config")
	slog.Info(b.image.Config.WorkingDir)

	j.Image = b.image

	return state
}
