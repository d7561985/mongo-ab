package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/d7561985/mongo-ab/cmd/mongo"
	"github.com/d7561985/mongo-ab/cmd/postgres"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Kill, os.Interrupt)

		<-ch

		log.Println("stop application")
		cancel()
	}()

	app := &cli.App{
		Name:  "mongo ab",
		Usage: "Compliance benchmark",
		Commands: []*cli.Command{
			mongo.New(),
			postgres.New(),
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
