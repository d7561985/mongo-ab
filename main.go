package main

import (
	"log"
	"os"

	"github.com/d7561985/mongo-ab/cmd/mongo"
	"github.com/d7561985/mongo-ab/cmd/postgres"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

func main() {
	app := &cli.App{
		Name:  "mongo ab",
		Usage: "Compliance benchmark",
		Commands: []*cli.Command{
			mongo.New(),
			postgres.New(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
