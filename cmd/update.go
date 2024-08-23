package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

var updateCmd = &cli.Command{
	Name:        "update",
	Description: "Update the server",
	Usage:       "Update the server",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "allow-experimental-builds", Aliases: []string{"e"}, Usage: "Allow experimental builds of Paper to be used"},
	},
	Before: func(c *cli.Context) error {
		fmt.Println("Updating the server...")
		return nil
	},
	After: func(c *cli.Context) error {
		fmt.Println("Update complete!")
		return nil
	},
	Action: func(c *cli.Context) error {
		fmt.Println("TODO: Update the server")
		return nil
	},
}

func init() {
	rootCmd.Commands = append(rootCmd.Commands, updateCmd)
}
