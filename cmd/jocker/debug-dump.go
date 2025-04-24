package jocker

import (
	"context"
	"log"
	"os"

	"github.com/google/go-jsonnet"
	"github.com/jocker-org/jocker/internal/parser"
	"github.com/moby/buildkit/client/llb"
)

func DebugDump() error {
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{JPaths: []string{"/lib/"}})

	jsonStr, err := vm.EvaluateFile("Jockerfile")
	if err != nil {
		log.Fatal(err)
	}

	j, err := parser.ParseJockerfile(jsonStr)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.TODO()
	state := j.ToLLB("all", ctx, nil)

	dt, err := state.Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		log.Fatal(err)
	}

	return llb.WriteTo(dt, os.Stdout)
}
