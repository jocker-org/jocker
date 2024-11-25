package parser

import (
	"encoding/json"

	"github.com/moby/buildkit/client/llb"
)

type BuildStep interface {
	ExecStep(llb.State) llb.State
}

type CopyStep struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type RunStep struct {
	Command string `json:"command"`
}

type BuildStage struct {
	Name string `json:"name"`
	From string `json:"from"`
	Steps *BuildSteps
}

type Jockerfile struct {
	Stages []BuildStage `json:"stages"`
}

type BuildSteps []BuildStep

func (steps *BuildSteps) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	for _, r := range raw {
		var obj map[string]interface{}
		err := json.Unmarshal(r, &obj)
		if err != nil {
			return err
		}
		
		stepType := ""
		if t, ok := obj["type"].(string); ok {
			stepType = t
		}

		var actual BuildStep
		switch stepType {
		case "COPY":
			actual = &CopyStep{}
		case "RUN":
			actual = &RunStep{}
		}
		err = json.Unmarshal(r, actual)
		if err != nil {
			return err
		}
		*steps = append(*steps, actual)		
	}
	return nil
}

func ParseJockerfile(jsonStr string) (*Jockerfile, error) {
	var j Jockerfile
	if err := json.Unmarshal([]byte(jsonStr), &j); err != nil {
		return nil, err
	}
	return &j, nil
}
