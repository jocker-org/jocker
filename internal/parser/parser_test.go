package parser

import (
	"os"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateJsonnetFile(t *testing.T) {
    jsonnetContent := `{
        name: 'test',
        value: 1 + 2
    }`
    tmpfile, err := os.CreateTemp("", "example.jsonnet")
    if err != nil {
        t.Fatalf("Failed to create temporary file: %v", err)
    }
    defer func() {
        if err := os.Remove(tmpfile.Name()); err != nil {
            t.Errorf("Failed to remove temporary file: %v", err)
        }
    }()

    _, err = tmpfile.Write([]byte(jsonnetContent))
    if err != nil {
        t.Fatalf("Failed to write to temporary file: %v", err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Errorf("Failed to close temporary file: %v", err)
    }

    result, err := EvaluateJsonnetFile(tmpfile.Name())
    if err != nil {
        t.Fatalf("EvaluateJsonnetFile returned an error: %v", err)
    }

    expectedOutput := `{
   "name": "test",
   "value": 3
}`

    assert.JSONEq(t, expectedOutput, result)
}

func TestCopyStep(t *testing.T) {
	b := &BuildContext{
		stages: map[string]llb.State{
			"stage1": llb.Scratch(),
		},
		state: llb.Scratch(),
		context: llb.Scratch(),
	}

	copyStep := &CopyStep{
		From:       "stage1",
		Source:     "/src/path",
		Destination: "/dest/path",
	}

	res, _ := copyStep.ExecStep(b)

	assert.NotEqual(t, res, llb.Scratch(), "BuildContext state should be updated")
}
