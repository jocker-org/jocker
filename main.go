package main

import (
	"log"

	cmd "github.com/jocker-org/jocker/cmd/jocker"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
