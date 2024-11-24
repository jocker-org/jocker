package parser

import (
	"encoding/json"
)

type Jockerfile struct {
	Image      string   `json:"image"`      // Base image
	Copy       []string `json:"copy"`       // Copy source:destination pairs
	Run        []string `json:"run"`        // Commands to execute during image build
	Cmd        []string `json:"cmd"`        // Command to run when container starts
	WorkDir    string   `json:"workdir"`    // Working directory inside the container
	Expose     []int    `json:"expose"`     // Ports to expose
	EntryPoint []string `json:"entrypoint"` // Entrypoint command
	Env        map[string]string `json:"env"` // Environment variables
}

func ParseJockerfile(jsonStr string) (*Jockerfile, error) {
	var j Jockerfile
	if err := json.Unmarshal([]byte(jsonStr), &j); err != nil {
		return nil, err
	}
	return &j, nil
}
