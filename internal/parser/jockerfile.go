package parser

import (
	"encoding/json"
)

type CopyStep struct {
	From        string `json:"from,omitempty"`
	Source      string `json:"src"`
	Destination string `json:"dst"`
}

type RunStep struct {
	Command string `json:"command"`
}

type WorkdirStep struct {
	Path string `json:"path"`
}

type BuildStage struct {
	Name string `json:"name"`
	From string `json:"from"`
	Steps *BuildSteps
	Cmd []string `json:"cmd"`
}

type Jockerfile struct {
	Stages []BuildStage `json:"stages"`
}

type BuildSteps []BuildStep

func (j *Jockerfile) Cmd() []string {
	if len(j.Stages) == 0 {
		return []string{}
	}

	return j.Stages[len(j.Stages)-1].Cmd
}

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
		case "WORKDIR":
			actual = &WorkdirStep{}
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
