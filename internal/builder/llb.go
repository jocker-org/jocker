package builder

import (
	"github.com/moby/buildkit/client/llb"
	"github.com/heph2/jocker/internal/parser"
	"strings"
	"fmt"
)

func JockerfileToLLB(j *parser.Jockerfile) llb.State {
	s := llb.Image(j.Image)
	// Not needed to pass the entire config,
	// just the Copy is enough
	if j.Copy != nil {
		s = JockerfileCopy(s, j)
	}

	if j.Run != nil {
		s = JockerfileRun(s, j)
	}
	return s
}

func JockerfileCopy(base llb.State, j *parser.Jockerfile) llb.State {
	st := llb.Local("context")
	// Here's the Copy is a List of pairs of this form
	// Foo:Bar
	for _, v := range j.Copy {
		parts := strings.Split(v, ":")
		base = base.File(llb.Copy(st, parts[0], parts[1]))
	}
	return base
}

func JockerfileRun(base llb.State, j *parser.Jockerfile) llb.State {
	for _, v := range j.Run {
		base = base.Run(shf(v)).Root()
	}
	return base
}

func JockerfileCmd(base llb.State, j *parser.Jockerfile) llb.State {
	for _, v := range j.Cmd {
		base = base.Run(shf(v)).Root()
	}
	return base
}

func shf(cmd string, v ...interface{}) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", fmt.Sprintf(cmd, v...)})
}
