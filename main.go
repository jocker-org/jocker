package main

import (
	"log"
	cmd "github.com/heph2/jocker/cmd/jocker"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
