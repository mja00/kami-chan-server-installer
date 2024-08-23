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
		// Just show the help message
		cli.ShowAppHelp(cCtx)
		return nil
	},
}

func init() {

}

func Run() {
	if err := rootCmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
