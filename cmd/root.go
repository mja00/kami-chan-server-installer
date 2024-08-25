package cmd

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var rootCmd = &cli.App{
	Name:  "kami-chan-server-installer",
	Usage: "Installer for Paper server",
	Action: func(cCtx *cli.Context) error {
		log.Println("No commands given. Installing server by default...")
		return setupCmd.Run(cCtx)
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "debug", Usage: "Enable debug mode"},
		&cli.BoolFlag{Name: "verbose", Usage: "Enable verbose mode"},
		&cli.StringFlag{Name: "server-dir", Usage: "Server directory", Value: "server"},
		&cli.BoolFlag{Name: "install-java-please", Usage: "This will install Java for you anyways on Linux"},
	},
	Version: Version,
}

var Version = "dev"
var Commit = "none"

func Run() {
	if err := rootCmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
