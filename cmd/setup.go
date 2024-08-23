package cmd

import (
	"fmt"
	"github.com/mja00/kami-chan-server-installer/paper"
	"github.com/urfave/cli/v2"
	"os"
)

var setupCmd = &cli.Command{
	Name:        "setup",
	Description: "Setup and install the server",
	Usage:       "Setup and install the server",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-prompts", Usage: "Skip setup prompts. This will only install Java and the jar file"},
		&cli.BoolFlag{Name: "accept-eula", Aliases: []string{"a"}, Usage: "Accept the EULA"},
		&cli.StringFlag{Name: "server-name", Aliases: []string{"n"}, Usage: "Server name", Value: "A server installed with Kami Chan Server Installer"},
		&cli.BoolFlag{Name: "whitelist", Aliases: []string{"w"}, Usage: "Enable whitelist"},
		&cli.BoolFlag{Name: "allow-experimental-builds", Aliases: []string{"e"}, Usage: "Allow experimental builds of Paper to be used"},
	},
	Before: func(c *cli.Context) error {
		fmt.Println("Setting up the server...")
		fmt.Printf("Server name: %s\n", c.String("server-name"))
		return nil
	},
	After: func(c *cli.Context) error {
		fmt.Println("Setup complete!")
		return nil
	},
	Action: func(c *cli.Context) error {
		// Create a server folder
		serverFolder := "server"
		if _, err := os.Stat(serverFolder); os.IsNotExist(err) {
			err := os.Mkdir(serverFolder, 0755)
			if err != nil {
				return err
			}
		}
		// First get our Paper API
		paperAPI := paper.NewPaperAPI()
		err := paperAPI.DownloadLatestBuild("paper", "server/paper.jar")
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.Commands = append(rootCmd.Commands, setupCmd)
}
