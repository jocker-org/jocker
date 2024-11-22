package parser

import (
	"encoding/json"
)

type Jockerfile struct {
	Image string `json:"image"`
	Copy  string `json:"copy"`
}

func ParseJockerfile(jsonStr string) (*Jockerfile, error) {
	var j Jockerfile
	if err := json.Unmarshal([]byte(jsonStr), &j); err != nil {
		return nil, err
	}
	return &j, nil
}
