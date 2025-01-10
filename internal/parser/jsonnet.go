package parser

import "github.com/google/go-jsonnet"
import "embed"
import "os"
import "fmt"

//go:embed libs/*
var std embed.FS

type LibImporter struct {
    fs embed.FS
}

func (l *LibImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
    // Try to read from embedded filesystem
    data, err := l.fs.ReadFile("libs/" + importedPath)
    if err != nil {
        // Fall back to filesystem if not found in embedded libs
        fsData, fsErr := os.ReadFile(importedPath)
        if fsErr != nil {
            return jsonnet.Contents{}, "", fmt.Errorf("could not find %s: %v", importedPath, err)
        }
        return jsonnet.MakeContents(string(fsData)), importedPath, nil
    }
    return jsonnet.MakeContents(string(data)), "libs/" + importedPath, nil
}

func EvaluateJsonnetFile(filePath string) (string, error) {
	vm := jsonnet.MakeVM()

	vm.Importer(&LibImporter{fs: std})

	return vm.EvaluateFile(filePath)
}

func EvaluateSnippet(filename, content string) (string, error) {
	vm := jsonnet.MakeVM()

	vm.Importer(&LibImporter{fs: std})

	return vm.EvaluateAnonymousSnippet(filename, content)
}
