package builder

import "github.com/moby/buildkit/client/llb"
import "github.com/heph2/jocker/internal/parser"

func JockerfileToLLB(j *parser.Jockerfile) llb.State {
	s := llb.Image(j.Image).File(llb.Copy(llb.Local("context"), j.Copy, j.Copy))
	return s
}
