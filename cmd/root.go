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
		fmt.Println("No commands given. Installing server by default...")
		return setupCmd.Run(cCtx)
	},
}

func init() {

}

func Run() {
	if err := rootCmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
