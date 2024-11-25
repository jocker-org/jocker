package parser

import (
	"fmt"

	"github.com/moby/buildkit/client/llb"
)

func (c *CopyStep) ExecStep(state llb.State) llb.State {
	st := llb.Local("context")
	state = state.File(llb.Copy(st, c.Source, c.Destination))
	return state
}

func (c *RunStep) ExecStep(state llb.State) llb.State {
	state = state.Run(shf(c.Command)).Root()
	return state
}

func shf(cmd string, v ...interface{}) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", fmt.Sprintf(cmd, v...)})
}

func (stage *BuildStage) ToLLB() llb.State {
	var state llb.State

	if stage.From == "scratch" {
		state = llb.Scratch()
	} else {
		state = llb.Image(stage.From)
	}

	for i := range *stage.Steps {
		state = (*stage.Steps)[i].ExecStep(state)
	}

	return state
}

func (j *Jockerfile) ToLLB() llb.State {
	var state llb.State

	for _, stage := range j.Stages {
		state = stage.ToLLB()
	}

	return state
}

// func JockerfileToLLB(j *parser.Jockerfile) llb.State {
// 	s := llb.Image(j.Image)
// 	// Not needed to pass the entire config,
// 	// just the Copy is enough
// 	if j.Copy != nil {
// 		s = JockerfileCopy(s, j)
// 	}

// 	if j.Run != nil {
// 		s = JockerfileRun(s, j)
// 	}
// 	return s
// }
