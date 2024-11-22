package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/google/go-jsonnet"
	"github.com/moby/buildkit/client/llb"
)

type Jockerfile struct {
	Image string `json:"image"`
	Copy string `json:"copy"`
}

func parseJSON(jsonStr string) (*Jockerfile, error) {
	var j Jockerfile
	if err := json.Unmarshal([]byte(jsonStr), &j); err != nil {
		return nil, err
	}
	return &j, nil
}

func Jockerfile2LLB(j *Jockerfile) (llb.State) {
	s := llb.Image(j.Image).File(llb.Copy(llb.Local("context"), j.Copy, j.Copy))
	return s
}

func main() {
	vm := jsonnet.MakeVM()
	jsonStr, err := vm.EvaluateFile("Jockerfile")
	if err != nil {
		log.Fatal(err)
	}
	j, err := parseJSON(jsonStr)
	if err != nil {
		log.Fatal(err)
	}

	dt, err := Jockerfile2LLB(j).Marshal(context.TODO(), llb.LinuxAmd64)
	if err != nil {
		panic(err)
	}
	llb.WriteTo(dt, os.Stdout)
}
