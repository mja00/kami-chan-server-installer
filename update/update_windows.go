package update

import (
	"github.com/goccy/go-json"
	"github.com/mja00/kami-chan-server-installer/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func GetUpdateURL() string {
	arch := utils.GetArch()
	var neededAsset string
	switch arch {
	case "x64":
		neededAsset = "kami-chan-server-installer_Windows_x86_64.tar.gz"
	default:
		return ""
	}
	githubAPI := "https://api.github.com/repos/mja00/kami-chan-server-installer/releases/latest"
	// Make the request
	resp, err := http.Get(githubAPI)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	// This will be a JSON object
	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}
	// Go through the assets and find one that matches kami-chan-server-installer_Darwin_{arch}.tar.gz
	for _, asset := range release.Assets {
		if asset.Name == neededAsset {
			return asset.BrowserDownloadUrl
		}
	}
	return ""
}

// Windows is a bit different, we most likely cannot overwrite the binary, as Windows will complain it's in use
// So we'll just extract the tar and place the new binary in the same directory as the old one with a new name
func InstallUpdate() error {
	log.Println("Unfortunately on Windows, we cannot directly update the application. We'll extract the new version and place it next to the old one.")
	// Untar the file, and then just replace the binary
	file, err := os.Open(filepath.Join("temp", "update.tar.gz"))
	if err != nil {
		return err
	}
	defer file.Close()
	utils.ExtractTarGz(file, filepath.Join("temp", "update"))
	_ = os.RemoveAll(filepath.Join("temp", "update.tar.gz"))
	// Move the binary to the main directory, overwriting the old one
	err = os.Rename(filepath.Join("temp", "update", "kami-chan-server-installer.exe"), "kami-chan-server-installer.exe")
	if err != nil {
		return err
	}
	// Remove the temp directory
	_ = os.RemoveAll(filepath.Join("temp", "update"))
	log.Println("Update complete! You can now run the new version of the binary.")
	return nil
}
