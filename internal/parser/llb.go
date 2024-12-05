package parser

import (
	"fmt"
	"log"

	"github.com/moby/buildkit/client/llb"
)

type BuildContext struct {
	stages map[string]llb.State
	state  llb.State
	context llb.State
}

type BuildStep interface {
	ExecStep(*BuildContext) (llb.State, error)
}

func (c *CopyStep) ExecStep(b *BuildContext) (llb.State, error) {
	st := b.context
	if c.From != "" {
		st = b.stages[c.From]
	}
	if c.Source == "" {
		return llb.State{}, fmt.Errorf("source path cannot be empty")
	}
	if c.Destination == "" {
		return llb.State{}, fmt.Errorf("destination path cannot be empty")
	}
	b.state = b.state.File(llb.Copy(st, c.Source, c.Destination))
	return b.state, nil
}

func (c *RunStep) ExecStep(b *BuildContext) (llb.State, error) {
	b.state = b.state.Run(shf(c.Command)).Root()
	return b.state, nil
}

func (c *WorkdirStep) ExecStep(b *BuildContext) (llb.State, error) {
	b.state = b.state.Dir(c.Path)
	return b.state, nil
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
		var err error
		b.state, err = (*stage.Steps)[i].ExecStep(b)
		if err != nil {
			log.Println("Failing to execute step", err)
		}
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
