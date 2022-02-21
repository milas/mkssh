package main

import (
	"log"
	"os"

	"github.com/milas/mkssh/internal/cli"
)

func main() {
	app, err := cli.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
