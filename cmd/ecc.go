package main

import (
	"Ecc/internal/importing"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var (
	DirPath = "../data"
)

func main() {
	var err error
	var model string
	dir := DirPath
	app := &cli.App{
		Name:  "Ecc",
		Usage: "Ecc is a tools for batch processing of excel data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "model",
				Aliases:     []string{"m"},
				Usage:       "The model of searching",
				Value:       "model",
				Destination: &model,
			},
			&cli.StringFlag{
				Name:        "dir",
				Aliases:     []string{"d"},
				Usage:       "Folder location of data files",
				Destination: &dir,
				Value:       DirPath,
			},
		},
		Action: func(c *cli.Context) error {
			importing.Load("../configs/cfg.yaml")
			importing.Handle(dir)
			return nil
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
