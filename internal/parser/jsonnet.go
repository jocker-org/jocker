package parser

import "github.com/google/go-jsonnet"

func EvaluateJsonnetFile(filePath string) (string, error) {
	vm := jsonnet.MakeVM()
	return vm.EvaluateFile(filePath)
}
