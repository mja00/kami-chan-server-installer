package cmd

import (
	"context"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/mja00/kami-chan-server-installer/cfg"
	"github.com/mja00/kami-chan-server-installer/minecraft"
	"github.com/mja00/kami-chan-server-installer/paper"
	"github.com/mja00/kami-chan-server-installer/utils"
	"github.com/pbnjay/memory"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	// This will be used for huh
	minecraftVersion  = "latest"
	serverName        = "A Minecraft Server"
	whitelist         = false
	acceptEULA        = false
	allowExperimental = false
)

var setupCmd = &cli.Command{
	Name:        "setup",
	Description: "Setup and install the server",
	Usage:       "Setup and install the server",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-prompts", Usage: "Skip setup prompts. This will only install Java and the jar file"},
		&cli.BoolFlag{Name: "accept-eula", Aliases: []string{"a"}, Usage: "Accept the EULA"},
		&cli.StringFlag{Name: "server-name", Aliases: []string{"n"}, Usage: "Server name", DefaultText: "A server installed with Kami Chan Server Installer", Value: "A Minecraft Server"},
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
		// Create the config file
		config := cfg.NewConfig()
		_ = config.Load(utils.GetServerFolder(".kami.json", c))
		// Add the config to the context
		c.Context = context.WithValue(c.Context, "config", config)
		// If debug pring all the flags
		if c.Bool("debug") {
			log.Println("Debug mode enabled")
			for _, flag := range c.FlagNames() {
				log.Printf("Flag: %s, Value: %s", flag, c.String(flag))
			}
		}
		for {
			err := prompt(c)
			if err != nil {
				// If the error is that we can't open a TTY, then just break from this loop. We'll use default values
				// error: huh: could not open a new TTY: open /dev/tty: no such device or address
				if strings.Contains(err.Error(), "could not open a new TTY") {
					break
				}
				// Otherwise, we'll just return the error
				return err
			}
			// Print out our settings
			fmt.Printf("Minecraft Version: %s\n", minecraftVersion)
			fmt.Printf("Allow Experimental Builds: %t\n", allowExperimental)
			fmt.Printf("Server Name: %s\n", serverName)
			fmt.Printf("Whitelist: %t\n", whitelist)
			// Ask if they want to save these settings
			var settingsGood bool
			_ = huh.NewConfirm().
				Title("Save these settings?").
				Description("Do you want to save these settings?").
				Value(&settingsGood).
				Run()
			if settingsGood {
				break
			}
		}
		return nil
	},
	After: func(c *cli.Context) error {
		log.Println("Setup complete!")
		// Save the config
		config := c.Context.Value("config").(*cfg.Config)
		err := config.Save(utils.GetServerFolder(".kami.json", c))
		if err != nil {
			return err
		}
		return nil
	},
	Action: func(c *cli.Context) error {
		// If they didn't already accept the EULA, then we need to prompt them to do so, unless we're skipping prompts, then error
		if !acceptEULA {
			// Error out
			return fmt.Errorf("you must accept the EULA to use this server")
		}
		// Grab the config
		config := c.Context.Value("config").(*cfg.Config)
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
		if minecraftVersion == "latest" {
			version, build, err := paperAPI.DownloadLatestBuild("paper", utils.GetServerFolder("paper.jar", c), allowExperimental)
			if err != nil {
				return err
			}
			config.SetMinecraftVersion(version)
			config.SetPaperBuild(strconv.Itoa(build))
		} else {
			// Download the specific version
			version, build, err := paperAPI.DownloadLatestBuildForVersion("paper", minecraftVersion, utils.GetServerFolder("paper.jar", c), allowExperimental)
			if err != nil {
				return err
			}
			config.SetMinecraftVersion(version)
			config.SetPaperBuild(strconv.Itoa(build))
		}
		// Check if the eula.txt file already exists and if the eula is already accepted
		eulaFile, err := os.ReadFile(utils.GetServerFolder("eula.txt", c))
		if err != nil {
			// If it's not found, then we need to write the eula.txt file
			err = os.WriteFile(utils.GetServerFolder("eula.txt", c), []byte("eula=true"), 0644)
			if err != nil {
				return err
			}
		}
		// If the eula is already accepted, then we're good
		if strings.Contains(string(eulaFile), "eula=true") {
			log.Println("EULA already accepted")
		} else {
			// Accept it
			err = os.WriteFile(utils.GetServerFolder("eula.txt", c), []byte("eula=true"), 0644)
			if err != nil {
				return err
			}
		}
		// Prompt the user for a MOTD/server name
		// Read our server.properties file
		// While this is a setup command, we're going to assume someone will run this accidentally, this will not wipe their config
		err = minecraft.ReadServerProperties(utils.GetServerFolder("server.properties", c))
		if err != nil {
			return err
		}
		viper.Set("motd", serverName)
		viper.Set("white-list", whitelist)
		// Write the server.properties file
		err = minecraft.WriteServerProperties(utils.GetServerFolder("server.properties", c))
		if err != nil {
			return err
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
		startScriptLocation := utils.GetStartScript(utils.GetServerFolder("start", c))
		// Ask the user if they want to start the server now or not
		var startServer bool
		_ = huh.NewConfirm().
			Title("Start the server?").
			Description("Do you want to start the server now?").
			Value(&startServer).
			Run()
		if startServer {
			// Start the server
			startScript := utils.GetStartScript("start")
			log.Println("Starting the server...")
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}
			// We need to change the working directory to the server folder
			err = os.Chdir(utils.GetServerFolder("", c))
			if err != nil {
				return err
			}
			err = utils.RunCommandAndPipeAllSTD(exec.Command("./"+startScript), true)
			if err != nil {
				return err
			}
			// cd back
			err = os.Chdir(pwd)
			// If we got here, the server was started and then stopped.
			// Just inform the user that to run the server again, they need to go into the server folder and run the start script
			log.Println("Server was successfully started!")
		}
		// They said no, just tell them how to run the server
		log.Println("To run the server, go into the server folder and run the start script.")
		log.Printf("The start script is located at: %s", startScriptLocation)
		return nil
	},
}

func init() {
	rootCmd.Commands = append(rootCmd.Commands, setupCmd)
}

func prompt(c *cli.Context) error {
	// Before we do anything, lets get some info from the user
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Minecraft Version").
				Description("What version of Minecraft do you want to use?").
				Value(&minecraftVersion).
				Placeholder(c.String("mc-version")).
				Validate(func(v string) error {
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
				}),
			// Allow experimental builds
			huh.NewConfirm().
				Title("Allow Experimental Builds").
				Description("Do you want to allow experimental builds of Paper to be used?").
				Value(&allowExperimental),
		),
		huh.NewGroup(
			// Server name
			huh.NewInput().
				Title("Server Name").
				Description("What do you want to name your server?").
				Value(&serverName).
				Placeholder("A Minecraft Server"),
			// Whitelist
			huh.NewConfirm().
				Title("Whitelist").
				Description("Do you want to enable the whitelist?").
				Value(&whitelist),
		),
		huh.NewGroup(
			// Accept EULA
			huh.NewConfirm().
				Title("Accept EULA").
				Description("Do you accept the Minecraft EULA?").
				Value(&acceptEULA),
		),
	)
	// Run the form, we'll use the results in the action
	err := form.Run()
	if err != nil {
		return err
	}
	return nil
}
