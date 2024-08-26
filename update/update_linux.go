package update

import (
	"github.com/goccy/go-json"
	"github.com/mja00/kami-chan-server-installer/utils"
	"net/http"
	"os"
	"path/filepath"
)

func GetUpdateURL() string {
	arch := utils.GetArch()
	var neededAsset string
	switch arch {
	case "x64":
		neededAsset = "kami-chan-server-installer_Linux_x86_64.tar.gz"
	case "aarch64":
		neededAsset = "kami-chan-server-installer_Linux_arm64.tar.gz"
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

func InstallUpdate() error {
	// Untar the file, and then just replace the binary
	file, err := os.Open(filepath.Join("temp", "update.tar.gz"))
	if err != nil {
		return err
	}
	defer file.Close()
	utils.ExtractTarGz(file, filepath.Join("temp", "update"))
	_ = os.RemoveAll(filepath.Join("temp", "update.tar.gz"))
	// Move the binary to the main directory, overwriting the old one
	err = os.Rename(filepath.Join("temp", "update", "kami-chan-server-installer"), "kami-chan-server-installer")
	if err != nil {
		return err
	}
	// Remove the temp directory
	_ = os.RemoveAll(filepath.Join("temp", "update"))
	return nil
}
