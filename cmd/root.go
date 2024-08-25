package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var rootCmd = &cli.App{
	Name:  "kami-chan-server-installer",
	Usage: "Installer for Paper server",
	Action: func(cCtx *cli.Context) error {
		log.Println("No commands given. Installing server by default...")
		return nil
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "debug", Usage: "Enable debug mode"},
		&cli.BoolFlag{Name: "verbose", Usage: "Enable verbose mode"},
		&cli.StringFlag{Name: "server-dir", Usage: "Server directory", Value: "server"},
		&cli.BoolFlag{Name: "install-java-please", Usage: "This will install Java for you anyways on Linux"},
	},
	Version:        Version,
	DefaultCommand: "setup",
	Before: func(cCtx *cli.Context) error {
		// Clear the terminal
		fmt.Println("\033[H\033[2J")
		return nil
	},
}

var Version = "dev"
var Commit = "none"

func Run() {
	if err := rootCmd.Run(os.Args); err != nil {
		log.Println(err)
		// Then hold the terminal open so the user can read the error
		fmt.Println("Press enter key to exit...")
		_, _ = fmt.Scanln()
	}
}
