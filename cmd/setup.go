package cmd

import (
	"fmt"
	"github.com/mja00/kami-chan-server-installer/minecraft"
	"github.com/mja00/kami-chan-server-installer/paper"
	"github.com/mja00/kami-chan-server-installer/utils"
	"github.com/pbnjay/memory"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

var setupCmd = &cli.Command{
	Name:        "setup",
	Description: "Setup and install the server",
	Usage:       "Setup and install the server",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-prompts", Usage: "Skip setup prompts. This will only install Java and the jar file"},
		&cli.BoolFlag{Name: "accept-eula", Aliases: []string{"a"}, Usage: "Accept the EULA"},
		&cli.StringFlag{Name: "server-name", Aliases: []string{"n"}, Usage: "Server name", DefaultText: "A server installed with Kami Chan Server Installer"},
		&cli.BoolFlag{Name: "whitelist", Aliases: []string{"w"}, Usage: "Enable whitelist"},
		&cli.BoolFlag{Name: "allow-experimental-builds", Aliases: []string{"e"}, Usage: "Allow experimental builds of Paper to be used"},
		&cli.StringFlag{Name: "mc-version", Aliases: []string{"m"}, Usage: "Minecraft version", Value: "latest", Action: func(ctx *cli.Context, v string) error {
			// If the version isn't latest then we need to check if it's valid format
			// Valid format is X.Y.Z or X.Y
			if v != "latest" {
				// Quick and dirty check to see if it's valid
				// If there is no "." thne it's definitely invalid
				if !strings.Contains(v, ".") {
					return fmt.Errorf("invalid Minecraft version: %s", v)
				}
				// If there is a "." then we need to make sure it's valid
				split := strings.Split(v, ".")
				if len(split) < 2 || len(split) > 3 {
					return fmt.Errorf("invalid Minecraft version: %s", v)
				}
				// So we know it's at least 2 parts, make sure all the parts are numbers
				for _, part := range split {
					if _, err := strconv.Atoi(part); err != nil {
						return fmt.Errorf("invalid Minecraft version: %s", v)
					}
				}
				// Good enough, we're good
			}
			return nil
		}},
	},
	Before: func(c *cli.Context) error {
		utils.PrintOSWarnings()
		log.Println("Setting up the server...")
		return nil
	},
	After: func(c *cli.Context) error {
		log.Println("Setup complete!")
		return nil
	},
	Action: func(c *cli.Context) error {
		// Check for Java
		log.Println("Checking for Java...")
		javaVersion, err := utils.GetJavaVersion()
		if err != nil {
			return err
		}
		log.Printf("Java version: %s\n", javaVersion.Version)
		// TODO: When Paper API v3 is released, we can check the recommended java version there. For now, we'll do a really shit check
		var requiredJavaVersion int
		if c.String("mc-version") == "latest" {
			// TODO: Don't hardcode this
			requiredJavaVersion = 21
		} else {
			requiredJavaVersion, err = utils.MCVersionToJavaMajor(c.String("mc-version"))
			if err != nil {
				return err
			}
		}
		if javaVersion.Major < requiredJavaVersion {
			log.Println("Java version is too low, downloading...")
			fileLoc, downloadErr := utils.DownloadJava(requiredJavaVersion, c)
			if downloadErr != nil {
				return downloadErr
			}
			log.Println("Installing Java...")
			err = utils.InstallJava(fileLoc, c)
			if err != nil {
				return err
			}
			// Re-verify the Java version
			javaVersion, err = utils.GetJavaVersion()
			if err != nil {
				return err
			}
			// If we're still too low then something went wrong, error and let the user figure it out
			if javaVersion.Major < requiredJavaVersion {
				return fmt.Errorf("java version must be at least %d", requiredJavaVersion)
			}
		}
		// Create a server folder
		log.Println("Downloading server files...")
		// Download our Paper jar
		paperAPI := paper.NewPaperAPI()
		if c.String("mc-version") == "latest" {
			err = paperAPI.DownloadLatestBuild("paper", utils.GetServerFolder("paper.jar", c), c.Bool("allow-experimental-builds"))
		} else {
			// Download the specific version
			err = paperAPI.DownloadLatestBuildForVersion("paper", c.String("mc-version"), utils.GetServerFolder("paper.jar", c), c.Bool("allow-experimental-builds"))
		}
		if err != nil {
			return err
		}
		// If they didn't already accept the EULA, then we need to prompt them to do so, unless we're skipping prompts, then error
		if !c.Bool("accept-eula") {
			// Check if the eula.txt file already exists and if the eula is already accepted
			eulaFile, err := os.ReadFile(utils.GetServerFolder("eula.txt", c))
			if err != nil {
				return err
			}
			var eulaAccepted bool
			if strings.Contains(string(eulaFile), "eula=true") {
				log.Println("EULA already accepted")
				eulaAccepted = true
			}
			// By default, we'll reject the EULA unless they say yes
			if !eulaAccepted {
				for {
					log.Println("You must accept the Minecraft EULA to use this server. You can read the EULA here: https://www.minecraft.net/en-us/eula Do you accept the EULA?")
					fmt.Print("Accept the EULA? [y/N] ")
					var response string
					_, err := fmt.Scanln(&response)
					if err != nil {
						return err
					}
					response = strings.ToLower(response)
					if response == "y" || response == "yes" {
						break
					} else {
						// No is default
						return fmt.Errorf("you must accept the EULA to use this server")
					}
				}
				// They accepted the EULA to get to this point. Write eula=true to the server/eula.txt file
				err = os.WriteFile(utils.GetServerFolder("eula.txt", c), []byte("eula=true"), 0644)
				if err != nil {
					return err
				}
			}
		}
		// Prompt the user for a MOTD/server name
		// Read our server.properties file
		// While this is a setup command, we're going to assume someone will run this accidentally, this will not wipe their config
		err = minecraft.ReadServerProperties(utils.GetServerFolder("server.properties", c))
		if err != nil {
			return err
		}
		if !c.Bool("skip-prompts") {
			// If they've already set a server name in the flags, then we'll use that
			if c.String("server-name") == "" {
				value := minecraft.PropertyPrompt("motd", "A Minecraft Server")
				viper.Set("motd", value)
			}
			// Same with the whitelist
			if !c.Bool("whitelist") {
				value := minecraft.ConfirmPrompt("whitelist")
				viper.Set("white-list", value)
			}
			// Write the server.properties file
			err = minecraft.WriteServerProperties(utils.GetServerFolder("server.properties", c))
			if err != nil {
				return err
			}
		}

		// Get the RAM of the machine
		totalRAM := float64(memory.TotalMemory())
		// At most we'll only ever set the script to use 10GB of RAM
		// Otherwise we'll use 75% of the total RAM
		ramAmount := int(math.Min(float64(10*1024*1024*1024), totalRAM*0.75))
		// Conver the amount to MB
		ramAmount = ramAmount / 1024 / 1024
		err = utils.WriteStartScript(utils.GetServerFolder("start", c), ramAmount, c)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.Commands = append(rootCmd.Commands, setupCmd)
}
