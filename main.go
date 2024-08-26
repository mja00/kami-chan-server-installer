package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mja00/kami-chan-server-installer/cmd"
	"github.com/mja00/kami-chan-server-installer/paper"
	"github.com/mja00/kami-chan-server-installer/update"
	"log"
)

var Version = "dev"
var Commit = "none"

func main() {
	// Clear the terminal
	fmt.Println("\033[H\033[2J")
	paper.Version = Version
	paper.Commit = Commit
	cmd.Version = Version
	cmd.Commit = Commit
	update.Version = Version
	update.Commit = Commit
	// Check for updates
	updateAvailable, err := update.CheckForUpdates()
	if err != nil {
		log.Println("Error checking for updates:", err)
	}
	if updateAvailable {
		color.Set(color.FgGreen)
		log.Println("An update is available!")
		url := update.GetUpdateURL()
		if url == "" {
			log.Println("Error downloading update")
		} else {
			//log.Println("Downloading update from:", url)
			//downloadErr := update.DownloadUpdate(url)
			//if downloadErr != nil {
			//	log.Println("Error downloading update:", downloadErr)
			//}
			//installErr := update.InstallUpdate()
			//if installErr != nil {
			//	log.Println("Error installing update:", installErr)
			//}
			//log.Println("Update complete!")
			//os.Exit(0)
			// For now just print the URL. We support downloading the update and installing it. Just it's a bit janky.
			log.Println("Download the update from:", url)
		}
		color.Unset()
	}
	cmd.Run()
}
