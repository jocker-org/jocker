package parser

import "github.com/google/go-jsonnet"

type ChainedImporter struct {
	ci *ContexImporter
	fi jsonnet.FileImporter
}

func NewChainedImporter(ci *ContexImporter, paths []string) *ChainedImporter {
	return &ChainedImporter{
		ci: ci,
		fi: jsonnet.FileImporter{
			JPaths: paths,
		},
	}
}

func (imp *ChainedImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	contents, foundHere, err := imp.fi.Import(importedFrom, importedPath)
	if err == nil {
		return contents, foundHere, err
	}

	return imp.ci.Import(importedFrom, importedPath)
}
