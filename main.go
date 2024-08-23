package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "kami-chan-server-installer",
		Usage: "Installer for Paper server",
		Action: func(cCtx *cli.Context) error {
			// Just show the help message
			cli.ShowAppHelp(cCtx)
			return nil
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:        "setup",
				Description: "Setup and install the server",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "accept-eula", Aliases: []string{"a"}, Usage: "Accept the EULA"},
					&cli.StringFlag{Name: "server-name", Aliases: []string{"n"}, Usage: "Server name"},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
