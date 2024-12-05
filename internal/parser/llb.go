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
	Evaluate(*BuildContext) error
}

func (c *CopyStep) Evaluate(b *BuildContext) error {
	st := b.context
	if c.From != "" {
		st = b.stages[c.From]
	}
	if c.Source == "" || c.Destination == "" {
		return fmt.Errorf("src or dst empty in copy operation")
	}
	b.state = b.state.File(llb.Copy(st, c.Source, c.Destination))
	return nil
}

func (c *RunStep) Evaluate(b *BuildContext) error {
	if c.Command == "" {
		return fmt.Errorf("empty command")
	}

	b.state = b.state.Run(shf(c.Command)).Root()
	return nil
}

func (c *WorkdirStep) Evaluate(b *BuildContext) error {
	b.state = b.state.Dir(c.Path)
	return nil
}

func shf(cmd string, v ...interface{}) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", fmt.Sprintf(cmd, v...)})
}

func (stage *BuildStage) ToLLB(b *BuildContext) (llb.State, error) {
	if stage.From == "scratch" {
		b.state = llb.Scratch()
	} else {
		b.state = llb.Image(stage.From)
	}

	for i := range *stage.Steps {
		log.Printf("building stage %#v\n", (*stage.Steps)[i])
		if err := (*stage.Steps)[i].Evaluate(b); err != nil {
			return b.state, err
		}
	}

	return b.state, nil
}

func (j *Jockerfile) ToLLB() (llb.State, error) {
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
		state, err := stage.ToLLB(&b)
		if err != nil {
			return state, err
		}
		b.stages[stage.Name] = state
	}

	return state, nil
}
