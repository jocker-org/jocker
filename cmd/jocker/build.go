package jocker

import (
	"log"

	"github.com/jocker-org/jocker/internal/parser"
	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
)

func Build() error {
	if err := grpcclient.RunFromEnvironment(appcontext.Context(), parser.Build); err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	return nil
}
