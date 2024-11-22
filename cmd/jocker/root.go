package jocker

import (
	"errors"
	"fmt"
	"os"
)

func Execute() error {
	if len(os.Args) < 2 {
		return errors.New("no command provided")
	}

	command := os.Args[1]
	// args := os.Args[2:]

	switch command {
	case "debug-dump":
		return DebugDump()
	case "help":
		return help()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func help() error {
	fmt.Println(`Usage:
  jocker <command> [arguments]

Commands:
  debug-dump  Print LLB definition to stdout
  help        Show this help message`)
	return nil
}
