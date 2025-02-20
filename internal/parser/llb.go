package parser

import (
	"fmt"
	"log"

	"github.com/moby/buildkit/client/llb"
)

type BuildContext struct {
	stages  map[string]llb.State
	state   llb.State
	context llb.State
}

type BuildStep interface {
	Evaluate(*BuildContext) llb.State
}

func (c *EnvStep) Evaluate(b *BuildContext) llb.State {
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

	b.state = b.state.Run(shf(c.Command)).Root()
	return b.state
}

func (c *WorkdirStep) Evaluate(b *BuildContext) llb.State {
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

func (stage *BuildStage) ToLLB(b *BuildContext) llb.State {
	if stage.From == "scratch" {
		b.state = llb.Scratch()
	} else {
		b.state = llb.Image(stage.From)
	}

	for i := range *stage.Steps {
		log.Printf("building stage %#v\n", (*stage.Steps)[i])
		b.state = (*stage.Steps)[i].Evaluate(b)
	}

	return b.state
}

func (j *Jockerfile) ToLLB() llb.State {
	b := BuildContext{
		stages: make(map[string]llb.State),
	}
	var state llb.State
	opts := []llb.LocalOption{
		llb.ExcludePatterns(j.Excludes),
	}

	b.context = llb.Local("context", opts...)

	for _, stage := range j.Stages {
		log.Println("building stage", stage.Name)

		state = stage.ToLLB(&b)
		b.stages[stage.Name] = state
	}

	return state
}
